-- Block pre-migration writers until both the backfill and compatibility trigger
-- are installed. SHARE ROW EXCLUSIVE conflicts with INSERT/UPDATE/DELETE's ROW
-- EXCLUSIVE lock while still permitting ordinary reads.
LOCK TABLE accounts IN SHARE ROW EXCLUSIVE MODE;

-- Reconcile existing user-owned connected accounts with owner-scoped Stripe
-- BYO subscription state. Preserve the operator's schedulability preference in
-- account metadata while forcing inactive BYO accounts out of schedulers. The
-- compatibility lock keeps older binaries safe during a rolling deployment.
WITH byo_accounts AS (
    SELECT a.id, a.owner_user_id
    FROM accounts a
    WHERE a.deleted_at IS NULL
      AND a.owner_user_id IS NOT NULL
)
UPDATE accounts a
SET schedulable = COALESCE((a.extra->>'byo_operational_schedulable')::boolean, a.schedulable),
    extra = COALESCE(a.extra, '{}'::jsonb) - 'byo_disabled_reason' - 'byo_operational_schedulable',
    updated_at = NOW()
FROM byo_accounts byo
WHERE a.id = byo.id
  AND COALESCE(a.extra->>'byo_disabled_reason', '') = 'subscription_inactive'
  AND EXISTS (
      SELECT 1
      FROM user_subscriptions us
      WHERE us.user_id = byo.owner_user_id
        AND us.deleted_at IS NULL
        AND us.status = 'active'
        AND us.expires_at > NOW()
        AND COALESCE(us.stripe_subscription_id, '') <> ''
  );

WITH byo_accounts AS (
    SELECT a.id, a.owner_user_id
    FROM accounts a
    WHERE a.deleted_at IS NULL
      AND a.owner_user_id IS NOT NULL
)
UPDATE accounts a
SET schedulable = FALSE,
    extra = jsonb_set(
        jsonb_set(
            COALESCE(a.extra, '{}'::jsonb),
            '{byo_operational_schedulable}',
            to_jsonb(a.schedulable),
            TRUE
        ),
        '{byo_disabled_reason}',
        to_jsonb('subscription_inactive'::text),
        TRUE
    ),
    updated_at = NOW()
FROM byo_accounts byo
WHERE a.id = byo.id
  AND COALESCE(a.extra->>'byo_disabled_reason', '') <> 'subscription_inactive'
  AND NOT EXISTS (
      SELECT 1
      FROM user_subscriptions us
      WHERE us.user_id = byo.owner_user_id
        AND us.deleted_at IS NULL
        AND us.status = 'active'
        AND us.expires_at > NOW()
        AND COALESCE(us.stripe_subscription_id, '') <> ''
  );

-- Existing scheduler workers understand this event and will discard snapshots
-- produced before the compatibility lock was written.
INSERT INTO scheduler_outbox (event_type, account_id, group_id, payload)
VALUES ('full_rebuild', NULL, NULL, NULL);

-- Protect the invariant while old and new application versions overlap during
-- a rolling deployment. Older binaries can still create owner-bound accounts
-- as schedulable or call SetSchedulable(true); the trigger records that desired
-- operational preference but keeps the effective value false until entitlement
-- is active. owner_user_id is reserved for user-connected (BYO) accounts.
CREATE OR REPLACE FUNCTION public.enforce_byo_subscription_compatibility_lock()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    entitled BOOLEAN;
    operational_schedulable BOOLEAN;
    previous_operational_schedulable BOOLEAN;
BEGIN
    IF NEW.owner_user_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT EXISTS (
        SELECT 1
        FROM user_subscriptions us
        WHERE us.user_id = NEW.owner_user_id
          AND us.deleted_at IS NULL
          AND us.status = 'active'
          AND us.expires_at > NOW()
          AND COALESCE(us.stripe_subscription_id, '') <> ''
    )
    INTO entitled;

    IF TG_OP = 'UPDATE' THEN
        previous_operational_schedulable := COALESCE(
            (OLD.extra->>'byo_operational_schedulable')::boolean,
            OLD.schedulable
        );
    END IF;

    IF entitled THEN
        IF TG_OP = 'UPDATE'
           AND COALESCE(OLD.extra->>'byo_disabled_reason', '') = 'subscription_inactive' THEN
            IF NEW.schedulable IS DISTINCT FROM OLD.schedulable THEN
                operational_schedulable := NEW.schedulable;
            ELSIF COALESCE(NEW.extra, '{}'::jsonb) ? 'byo_operational_schedulable'
                  AND (NEW.extra->>'byo_operational_schedulable') IS DISTINCT FROM
                      (OLD.extra->>'byo_operational_schedulable') THEN
                operational_schedulable := (NEW.extra->>'byo_operational_schedulable')::boolean;
            ELSE
                operational_schedulable := previous_operational_schedulable;
            END IF;
            NEW.schedulable := operational_schedulable;
            NEW.extra := COALESCE(NEW.extra, '{}'::jsonb)
                - 'byo_disabled_reason'
                - 'byo_operational_schedulable';
        ELSIF COALESCE(NEW.extra->>'byo_disabled_reason', '') = 'subscription_inactive' THEN
            NEW.schedulable := COALESCE(
                (NEW.extra->>'byo_operational_schedulable')::boolean,
                NEW.schedulable
            );
            NEW.extra := COALESCE(NEW.extra, '{}'::jsonb)
                - 'byo_disabled_reason'
                - 'byo_operational_schedulable';
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE'
       AND COALESCE(OLD.extra->>'byo_disabled_reason', '') = 'subscription_inactive' THEN
        IF NEW.schedulable IS DISTINCT FROM OLD.schedulable THEN
            operational_schedulable := NEW.schedulable;
        ELSIF COALESCE(NEW.extra, '{}'::jsonb) ? 'byo_operational_schedulable'
              AND (NEW.extra->>'byo_operational_schedulable') IS DISTINCT FROM
                  (OLD.extra->>'byo_operational_schedulable') THEN
            -- SetSchedulable explicitly stores no-op false requests here,
            -- because the effective locked value is already false.
            operational_schedulable := (NEW.extra->>'byo_operational_schedulable')::boolean;
        ELSE
            -- Full account updates always mention schedulable. Preserve the old
            -- preference when that effective field and its preference did not
            -- actually change.
            operational_schedulable := previous_operational_schedulable;
        END IF;
    ELSIF COALESCE(NEW.extra, '{}'::jsonb) ? 'byo_operational_schedulable' THEN
        -- New writers pre-seed the operator preference before forcing the lock.
        operational_schedulable := (NEW.extra->>'byo_operational_schedulable')::boolean;
    ELSE
        -- Covers owner-bound account inserts from pre-migration binaries.
        operational_schedulable := NEW.schedulable;
    END IF;

    NEW.extra := jsonb_set(
        jsonb_set(
            COALESCE(NEW.extra, '{}'::jsonb),
            '{byo_operational_schedulable}',
            to_jsonb(operational_schedulable),
            TRUE
        ),
        '{byo_disabled_reason}',
        to_jsonb('subscription_inactive'::text),
        TRUE
    );
    NEW.schedulable := FALSE;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_enforce_byo_subscription_compatibility_lock ON accounts;
CREATE TRIGGER trg_enforce_byo_subscription_compatibility_lock
BEFORE INSERT OR UPDATE OF schedulable, owner_user_id ON accounts
FOR EACH ROW
EXECUTE FUNCTION public.enforce_byo_subscription_compatibility_lock();
