# Schedule tables

```sql
CREATE TABLE IF NOT EXISTS scheduler.provider_match (
    name_match TEXT NOT NULL,
    provider_id UUID NOT NULL REFERENCES scheduler.providers (id),
    PRIMARY KEY (name_match, provider_id)
);

CREATE TABLE IF NOT EXISTS scheduler.providers_v2_migration (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    calendar_id uuid,
    external_id text,
    external_name character varying(100),
    location_id uuid NOT NULL,
    source_tenant_id uuid,
    is_integrated boolean DEFAULT false NOT NULL,
    hidden boolean DEFAULT false,
    first_name character varying(50),
    middle_name character varying(50),
    last_name character varying(50),
    display_name character varying(100),
    profile jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    origin text,
    source_id uuid,
    CONSTRAINT providers_check CHECK ((((is_integrated IS TRUE) AND (external_id IS NOT NULL)) OR ((is_integrated IS FALSE) AND (external_id IS NULL))))
);

CREATE TABLE scheduler.mappings
(
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    mapping_type        text                           NOT NULL,
    location_id         uuid                           NOT NULL, -- in multi case this will be the child location_id, not the parent location_id
    provider_id         uuid,
    operatory_id        uuid,
    appointment_type_id uuid,
    created_at          timestamptz      DEFAULT NOW() NOT NULL,
    FOREIGN KEY (provider_id) REFERENCES scheduler.providers (id) ON DELETE CASCADE,
    FOREIGN KEY (operatory_id) REFERENCES scheduler.operatories (id) ON DELETE CASCADE,
    FOREIGN KEY (appointment_type_id) REFERENCES scheduler.appointment_types (id) ON DELETE CASCADE
);

CREATE TABLE scheduler.booking_submissions
(
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id            uuid                           NOT NULL,
    attendee_id            uuid,                                    -- future for OTP based verification, if set means existing user (mostly patient id)
    appointment_type_id    uuid                           NOT NULL,
    provider_id            uuid,
    requested_slots        varchar[],
    slot_duration          int,

    booking_source         varchar(64)                    NOT NULL, -- Customer Website,Google Business Profile etc
    first_name             varchar(64)                    NOT NULL,
    middle_name            varchar(64),
    last_name              varchar(64),
    birthdate              date,
    email                  varchar(256)                   NOT NULL,
    phone_number           varchar(64)                    NOT NULL,
    address_info           jsonb,                                   -- address, address2, city, state, postcode, country
    form_metadata          jsonb,                                   -- other custom key value
    payment_transaction_id varchar,
    note                   varchar,

    reviewed_by            uuid,
    reviewed_at            timestamptz,
    reviewed_status        varchar(64)                    NOT NULL, -- "UnknownStatus","Pending", "Accepted","Rejected"

    created_at             timestamptz      DEFAULT NOW() NOT NULL,
    updated_at             timestamptz,
    FOREIGN KEY (provider_id) REFERENCES scheduler.providers (id),
    FOREIGN KEY (appointment_type_id) REFERENCES scheduler.appointment_types (id)
);

CREATE TABLE scheduler.providers
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid,                                    -- weave's user ID
    calendar_id      uuid                           NOT NULL, -- calendar_id of the provider
    external_id      text,                                    -- synced provider id [integrated location]
    external_name    varchar(100),                            -- synced provider name [integrated location]
    location_id      uuid                           NOT NULL,
    source_tenant_id uuid                           NOT NULL, -- office_id
    is_integrated    boolean          DEFAULT false NOT NULL,
    hidden           boolean          DEFAULT false,
    first_name       varchar(50),
    middle_name      varchar(50),
    last_name        varchar(50),
    display_name     varchar(100),
    profile          jsonb,
    created_at       timestamptz      DEFAULT NOW() NOT NULL,
    updated_at       timestamptz,
    deleted_at       timestamptz,
    FOREIGN KEY (calendar_id) REFERENCES scheduler.calendar (id),
    CHECK (
        (is_integrated IS TRUE AND external_id IS NOT NULL)
            OR
        (is_integrated IS FALSE AND external_id IS NULL)
        )                                                     -- external_id is required if integrated is true
);

CREATE TABLE scheduler.operatories
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id      uuid                           NOT NULL,
    source_tenant_id uuid                           NOT NULL, -- office_id
    external_id      text,                                    -- synced operatory id [integrated location]
    external_name    varchar(100),                            -- synced operatory name  [integrated location]
    display_name     varchar(100),
    is_integrated    boolean          DEFAULT false NOT NULL,
    hidden           boolean          DEFAULT false,
    details          jsonb,
    created_at       timestamptz      DEFAULT NOW() NOT NULL,
    updated_at       timestamptz,
    deleted_at       timestamptz,
    CHECK (
        (is_integrated IS TRUE AND external_id IS NOT NULL AND external_name IS NOT NULL)
            OR
        (is_integrated IS FALSE AND external_id IS NULL AND external_name IS NULL)
        )
);

CREATE TABLE scheduler.appointment_types_practitioners
(
    location_id         uuid                      NOT NULL, -- in multi case this will be the child location_id, not the parent location_id
    practitioner_id     uuid                      NOT NULL,
    appointment_type_id uuid                      NOT NULL,
    created_at          timestamptz DEFAULT NOW() NOT NULL,
    PRIMARY KEY (location_id, practitioner_id, appointment_type_id),
    FOREIGN KEY (practitioner_id) REFERENCES scheduler.practitioners (id) ON DELETE CASCADE,
    FOREIGN KEY (appointment_type_id) REFERENCES scheduler.appointment_types (id) ON DELETE CASCADE
);

CREATE TABLE scheduler.appointment_types
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id      uuid                           NOT NULL,
    source_tenant_id uuid                           NOT NULL, -- office id
    external_id      text,                                    -- synced appointment type id [integrated location]
    external_name    varchar(100),                            -- synced appointment type name [integrated location]
    display_name     varchar(100),
    description      text,
    is_integrated    boolean          DEFAULT false NOT NULL,
    hidden           boolean          DEFAULT false,
    details          jsonb,
    created_at       timestamptz      DEFAULT NOW() NOT NULL,
    updated_at       timestamptz,
    deleted_at       timestamptz,
    CHECK (
        (is_integrated IS TRUE AND external_id IS NOT NULL AND external_name IS NOT NULL)
            OR
        (is_integrated IS FALSE AND external_id IS NULL AND external_name IS NULL)
        )
);


create table scheduler.schedule
(
    id               uuid                      NOT NULL, -- location_id, provider_id or operatory_id
    location_id      uuid                      NOT NULL,
    type             varchar(50)               NOT NULL, -- in 'LOCATION', 'PROVIDER' or 'OPERATORY' or 'HOLIDAY' etc
    recurrence_rules jsonb                     NOT NULL,
    is_integrated    boolean     DEFAULT false NOT NULL,
    created_at       timestamptz DEFAULT NOW() NOT NULL,
    updated_at       timestamptz,
    PRIMARY KEY (id, location_id)
);

CREATE TABLE scheduler.practitioners
(
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid,                                    -- weave's user ID
    calendar_id      uuid                           NOT NULL, -- calendar_id of the practitioner
    external_id      text,                                    -- synced practitioner id [integrated location]
    external_name    varchar(100),                            -- synced practitioner name [integrated location]
    location_id      uuid                           NOT NULL,
    source_tenant_id uuid                           NOT NULL, -- office_id
    is_integrated    boolean          DEFAULT false NOT NULL,
    hidden           boolean          DEFAULT false,
    first_name       varchar(50),
    middle_name      varchar(50),
    last_name        varchar(50),
    display_name     varchar(100),
    profile          jsonb,
    created_at       timestamptz      DEFAULT NOW() NOT NULL,
    updated_at       timestamptz,
    deleted_at       timestamptz,
    FOREIGN KEY (calendar_id) REFERENCES scheduler.calendar (id),
    CHECK (
        (is_integrated IS TRUE AND external_id IS NOT NULL)
            OR
        (is_integrated IS FALSE AND external_id IS NULL)
        )                                                     -- external_id is required if integrated is true
);

CREATE TABLE scheduler.calendars_events
(
    id                    uuid                      NOT NULL DEFAULT gen_random_uuid(),
    location_id           uuid                      NOT NULL,
    source_tenant_id      uuid                      NOT NULL,                           -- office_id
    organizer_calendar_id uuid                      NOT NULL,                           -- calendar_id of the organizer
    organizer_id          uuid                      NOT NULL,                           -- The organizer of the event. (mostly internal weave provider_id)
    location_type         varchar(50)               NOT NULL,                           -- room,operatory,virtual, etc
    type                  varchar(50)               NOT NULL,                           -- APPOINTMENT, TODOs, REMINDER, and BUSY]
    operatory_id          uuid,                                                         -- operatory_id for the event
    reference_type_id     uuid,                                                         -- mostly weave appointment type id.
    is_integrated         boolean     DEFAULT false NOT NULL,
    reference_id          uuid                      NOT NULL DEFAULT gen_random_uuid(), -- appointment_id, task_id, etc. for synced events like appointment ids will be generated by the service.
    attendee_id           uuid                      NOT NULL,                           -- internal weave id of the person having the appointment
    attendee_status       varchar(20)               NOT NULL,                           -- ATTEMPTED, CONFIRMED, UNCONFIRMED, CANCELED, COMPLETED, NO_SHOW
    title                 varchar(100),
    start_date            date                      NOT NULL,
    start_time            timestamp                 NOT NULL,
    start_timezone        varchar(50),                                                  -- location timezone or organizer timezone
    end_date              date                      NOT NULL,
    end_time              timestamp                 NOT NULL,
    end_timezone          varchar(50),
    details               jsonb,                                                        -- store flexible event details without needing to alter the table structure for each event type.
    recurring             boolean     DEFAULT false NOT NULL,
    recurrence_rule       text,                                                         -- iCal RFC 5545
    created_at            timestamptz DEFAULT NOW() NOT NULL,
    updated_at            timestamptz,
    deleted_at            timestamptz,
    FOREIGN KEY (organizer_calendar_id) REFERENCES scheduler.calendar (id),
    CHECK (
        (recurring IS TRUE AND recurrence_rule IS NOT NULL)
            OR
        (recurring IS FALSE AND recurrence_rule IS NULL)
        )
) PARTITION BY RANGE (start_date);

CREATE TABLE scheduler.calendar
(
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id       uuid, -- weave's user ID, storing here so that non-providers can also have calendars
    location_id    uuid                           NOT NULL,
    title          varchar(100),
    description    text,
    timezone       varchar(50)                    NOT NULL,
    view_summaries boolean,
    created_at     timestamptz      DEFAULT NOW() NOT NULL,
    updated_at     timestamptz,
    deleted_at     timestamptz
);

CREATE TABLE IF NOT EXISTS scheduler.settings
(
    location_id     UUID PRIMARY KEY,
    updated_by      UUID,                  -- user who created/updated the record
    lead_duration   integer     DEFAULT 0, -- lead time in minutes before the appointment, 0 means no lead time. [Booking Request]
    booking_deposit integer     DEFAULT 0, -- booking deposit amount in dollar, 0 means no deposit required. [Booking Request]
    created_at      timestamptz DEFAULT NOW() NOT NULL,
    updated_at      timestamptz
);
```