# Persons tables

Table "public.account_relationship"
      Column         |           Type           | Collation | Nullable |      Default
---------------------+--------------------------+-----------+----------+-------------------
 org_id              | uuid                     |           | not null |
 dataset_id          | uuid                     |           |          |
 source_tenant_id    | uuid                     |           |          |
 source_id           | uuid                     |           | not null |
 id                  | uuid                     |           | not null |
 account_relation_pmid| character varying(128) |           | not null |
 person_id           | uuid                     |           | not null |
 account_id          | uuid                     |           | not null |
 account_pmid        | character varying(128)   |           | not null |
 relationship        | character varying(64)    |           | not null |
 created_at          | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at         | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at          | timestamp with time zone |           |          |
Indexes:
    "account_relationship_pkey" PRIMARY KEY, btree (id)
    "account_relationship_deleted_at" btree (deleted_at) WHERE deleted_at IS NOT NULL
    "account_relationship_person_3x" btree (person_id) WHERE dataset_id IS NULL OR source_tenant_id IS NULL
    "account_relationship_org_id_2x" btree (org_id, person_id, id) WHERE deleted_at IS NULL
    "account_relationship_org_merge" btree (org_id, source_id)

Table "public.address"
      Column        |           Type           | Collation | Nullable |      Default
--------------------+--------------------------+-----------+----------+-------------------
 org_id             | uuid                     |           | not null |
 dataset_id         | uuid                     |           | not null |
 source_tenant_id   | uuid                     |           | not null |
 source_id          | uuid                     |           | not null |
 id                 | uuid                     |           | not null |
 person_id          | uuid                     |           | not null |
 address_lines      | character varying(256)   |           |          |
 city               | character varying(64)    |           |          |
 state              | character varying(64)    |           |          |
 postal_code        | character varying(64)    |           |          |
 created_at         | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at        | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at         | timestamp with time zone |           |          |
Indexes:
    "address_pkey" PRIMARY KEY, btree (id)
    "address_person_id" UNIQUE, btree (person_id)
    "address_deleted_at" btree (deleted_at) WHERE deleted_at IS NOT NULL
    "address_org_merge" btree (org_id, source_id)

Table "public.client_location"
      Column         |           Type           | Collation | Nullable |      Default
---------------------+--------------------------+-----------+----------+-------------------
 org_id              | uuid                     |           | not null |
 source_id           | uuid                     |           | not null |
 source_tenant_id    | uuid                     |           |          |
 id                  | uuid                     |           | not null |
 external_id         | character varying(64)    |           | not null |
 name                | character varying(128)   |           | not null |
 shortname           | character varying(64)    |           |          |
 address             | character varying(128)   |           |          |
 address2            | character varying(128)   |           |          |
 city                | character varying(64)    |           |          |
 state               | character varying(64)    |           |          |
 postal_code         | character varying(64)    |           |          |
 external_inactive   | boolean                  |           | not null |
 client_location_type| character varying(64)    |           |          |
 created_at          | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at         | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at          | timestamp with time zone |           |          |
Indexes:
    "client_location_pkey" PRIMARY KEY, btree (id)
    "client_location_source_id_idx" btree (source_id) INCLUDE (source_tenant_id)
    "client_location_org_id_3x_idx" btree (org_id, source_id, source_tenant_id) WHERE deleted_at IS NULL

Table "public.contact_info"
       Column          |           Type           | Collation | Nullable |      Default
-----------------------+--------------------------+-----------+----------+-------------------
 org_id                | uuid                     |           | not null |
 dataset_id            | uuid                     |           |          |
 source_tenant_id      | uuid                     |           |          |
 source_id             | uuid                     |           | not null |
 id                    | uuid                     |           | not null |
 contact_info_pmid     | character varying(256)   |           | not null |
 person_id             | uuid                     |           | not null |
 type                  | character varying(64)    |           | not null |
 type_pm               | character varying(64)    |           |          |
 destination           | character varying(256)   |           |          |
 priority              | integer                  |           | not null |
 created_at            | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at           | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at            | timestamp with time zone |           |          |
 normalized_destination| character varying(256)   |           |          |
Indexes:
    "contact_info_pkey" PRIMARY KEY, btree (id)
    "contact_info_person_5x" btree (person_id, org_id, type, priority, dataset_id) WHERE deleted_at IS NULL
    "contact_info_person_3x" btree (person_id) WHERE dataset_id IS NULL OR source_tenant_id IS NULL
    "contact_info_org_id_2x" btree (org_id, destination) WHERE deleted_at IS NULL
    "contact_info_deleted_at" btree (deleted_at) WHERE deleted_at IS NOT NULL
    "contact_info_last_four_search_idx" btree (last4(normalized_destination::text, type::text))
    "contact_info_compliance_dest" btree (destination, org_id)
    "contact_info_normalized_dest_3x" btree (org_id, normalized_destination, type, person_id) WHERE deleted_at IS NULL
    "contact_info_org_merge" btree (org_id, source_id)
    "contact_info_modified_at_temp" btree (modified_at, id)
    "contact_info_org_id_dest_3x" btree (org_id, destination, type) INCLUDE (person_id) WHERE deleted_at IS NULL
    "contact_info_org_id_dest_3bx" btree (org_id, destination varchar_pattern_ops, type) INCLUDE (person_id) WHERE deleted_at IS NULL
    "contact_info_org_id_last_four_search_idx" btree (org_id, last4(normalized_destination::text, type::text) text_pattern_ops) WHERE deleted_at IS NULL
    "contact_info_backfill_temp" btree (modified_at, id) INCLUDE (destination) WHERE type::text = ANY (ARRAY['mobile','home','work']) AND destination IS NOT NULL

Table "public.dataset"
      Column       |           Type           | Collation | Nullable |      Default
-------------------+--------------------------+-----------+----------+-------------------
 org_id            | uuid                     |           | not null |
 dataset_id        | uuid                     |           | not null |
 source_tenant_id  | uuid                     |           | not null |
 source_id         | uuid                     |           | not null |
 synced_at         | timestamp with time zone |           |          |
 created_at        | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at       | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at        | timestamp with time zone |           |          |
Indexes:
    "dataset_pkey" PRIMARY KEY, btree (dataset_id)
    "dataset_synced_at_created_at" btree (synced_at, created_at)

Table "public.location_access_rule"
      Column        |           Type           | Collation | Nullable |      Default
--------------------+--------------------------+-----------+----------+-------------------
 locationaccessid   | uuid                     |           | not null |
 locationid         | uuid                     |           |          |
 sourceid           | uuid                     |           |          |
 clientlocationid   | uuid                     |           |          |
 accesstype         | character varying(64)    |           |          |
 createdat          | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deletedat          | timestamp with time zone |           |          |
 modifiedat         | timestamp with time zone |           |          | now()
 userid             | uuid                     |           |          |
Indexes:
    "location_access_rule_pkey" PRIMARY KEY, btree (locationaccessid)
    "location_access_rule_clientlocationid_idx" btree (clientlocationid) INCLUDE (locationid)
    "location_access_rule_sourceid_idx" btree (sourceid) INCLUDE (locationid, clientlocationid)
    "location_access_rule_locationid_idx" btree (locationid) INCLUDE (clientlocationid, sourceid)

Table "public.person"
      Column       |           Type           | Collation | Nullable |      Default
-------------------+--------------------------+-----------+----------+-------------------
 org_id            | uuid                     |           | not null |
 dataset_id        | uuid                     |           | not null |
 source_tenant_id  | uuid                     |           | not null |
 source_id         | uuid                     |           | not null |
 id                | uuid                     |           | not null |
 person_pmid       | character varying(64)    |           | not null |
 person_display_pmid| character varying(64)   |           |          |
 household_id      | uuid                     |           | not null |
 household_pmid    | character varying(64)    |           | not null |
 is_guardian       | boolean                  |           | not null |
 first_name        | character varying(64)    |           |          |
 last_name         | character varying(64)    |           |          |
 preferred_name    | character varying(64)    |           |          |
 status            | character varying(64)    |           |          |
 gender            | character varying(64)    |           |          |
 birthdate         | date                     |           |          |
 notes             | character varying(255)   |           |          |
 entry_date        | timestamp with time zone |           |          |
 created_at        | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at       | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at        | timestamp with time zone |           |          |
Indexes:
    "person_pkey" PRIMARY KEY, btree (id)
    "person_status_4x" btree (org_id, status, dataset_id, birthdate) WHERE deleted_at IS NULL
    "person_household_id_3x" btree (org_id, household_id, dataset_id) WHERE deleted_at IS NULL
    "person_deleted_at" btree (deleted_at) WHERE deleted_at IS NOT NULL
    "person_deleted_at_v2" btree (deleted_at) INCLUDE (id) WHERE deleted_at IS NOT NULL
    "person_org_household_4x" btree (org_id, source_tenant_id, household_id, id) WHERE deleted_at IS NULL
    "person_org_id_source_id_2x" btree (org_id, source_id, source_tenant_id, id)
    "person_org_createdat" btree (org_id, source_tenant_id, created_at, id) WHERE deleted_at IS NULL
    "person_org_household_5x" btree (org_id, source_tenant_id, household_id, created_at, id) WHERE deleted_at IS NULL
    "person_backfill_pagination" btree (modified_at, id)
    "person_org_id_3x" btree (org_id, source_tenant_id, id) WHERE deleted_at IS NULL
    "person_page_lfid" btree (org_id, source_tenant_id, COALESCE(last_name, 'ZZZZZ'::character varying), COALESCE(first_name, 'ZZZZZ'::character varying), id) WHERE deleted_at IS NULL
    "person_nopage_lfid" btree (org_id, source_tenant_id, last_name, first_name, id) WHERE deleted_at IS NULL
    "person_id_3x" btree (id, org_id, source_tenant_id) WHERE deleted_at IS NULL
Statistics objects:
    "person_org_sourcet_stat" (dependencies) ON source_tenant_id, org_id

Table "public.person_search"
              Column               |           Type           | Collation | Nullable |      Default
-----------------------------------+--------------------------+-----------+----------+-------------------
 org_id                            | uuid                     |           | not null |
 dataset_id                        | uuid                     |           | not null |
 source_tenant_id                  | uuid                     |           | not null |
 source_id                         | uuid                     |           | not null |
 person_id                         | uuid                     |           | not null |
 normalized_first_name_last_name   | character varying(256)   |           | not null |
 normalized_last_name_first_name   | character varying(256)   |           | not null |
 normalized_preferred_name_last_name| character varying(256)  |           | not null |
 created_at                        | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at                       | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 normalized_display_pm_id          | character varying(256)   |           |          |
Indexes:
    "person_search_pkey" PRIMARY KEY, btree (person_id)
    "person_search_normalized_first_name_last_name_3x" btree (org_id, dataset_id, normalized_first_name_last_name) INCLUDE (person_id)
    "person_search_normalized_last_name_first_name_3x" btree (org_id, dataset_id, normalized_last_name_first_name) INCLUDE (person_id)
    "person_search_normalized_preferred_name_last_name_3x" btree (org_id, dataset_id, normalized_preferred_name_last_name) INCLUDE (person_id)
    "persons_search_normalized_display_pm_id" btree (org_id, dataset_id, normalized_display_pm_id) INCLUDE (person_id)
    "persons_search_normalized_display_pm_id_2" btree (org_id, normalized_display_pm_id)
    "persons_search_normalized_display_pm_id_3x" btree (org_id, source_tenant_id, normalized_display_pm_id varchar_pattern_ops) INCLUDE (person_id)
    "person_search_org_merge" btree (org_id, source_id)
    "person_search_backfill_pagination" btree (modified_at, person_id)
    "idx_person_search_empty_normalized_2" btree (modified_at, person_id) WHERE normalized_preferred_name_last_name::text = ''::text
    "person_search_normalized_first_name_last_name_3ax" btree (org_id, source_tenant_id, normalized_first_name_last_name varchar_pattern_ops) INCLUDE (person_id)
    "person_search_normalized_last_name_first_name_3bx" btree (org_id, source_tenant_id, normalized_last_name_first_name COLLATE "C") INCLUDE (person_id)
    "person_search_normalized_last_name_first_name_3ax" btree (org_id, source_tenant_id, normalized_last_name_first_name varchar_pattern_ops) INCLUDE (person_id)
    "person_search_normalized_preferred_name_last_name_3ax" btree (org_id, source_tenant_id, normalized_preferred_name_last_name varchar_pattern_ops) INCLUDE (person_id)
    "person_search_normalized_first_name_last_name_3xd" btree (org_id, source_tenant_id, normalized_first_name_last_name varchar_pattern_ops) INCLUDE (person_id)
    "person_search_normalized_last_name_first_name_3xd" btree (org_id, source_tenant_id, normalized_last_name_first_name varchar_pattern_ops) INCLUDE (person_id)
Dropped by later migration:
    "persons_search_normalized_display_pm_id_st"

Table "public.preference"
      Column       |           Type           | Collation | Nullable |      Default
-------------------+--------------------------+-----------+----------+-------------------
 org_id            | uuid                     |           | not null |
 dataset_id        | uuid                     |           | not null |
 source_tenant_id  | uuid                     |           | not null |
 source_id         | uuid                     |           | not null |
 id                | uuid                     |           | not null |
 person_id         | uuid                     |           | not null |
 lang              | integer                  |           | not null |
 practitioner      | character varying(64)    |           |          |
 created_at        | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at       | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 deleted_at        | timestamp with time zone |           |          |
Indexes:
    "preference_pkey" PRIMARY KEY, btree (id)
    "preference_person_id" UNIQUE, btree (person_id)
    "preference_deleted_at" btree (deleted_at) WHERE deleted_at IS NOT NULL
    "preference_org_merge" btree (org_id, source_id)
    "preference_person_id_idx" btree (person_id) INCLUDE (lang, practitioner)

Table "public.primary_contact"
      Column      |           Type           | Collation | Nullable |      Default
------------------+--------------------------+-----------+----------+-------------------
 location_id      | uuid                     |           | not null |
 id               | uuid                     |           | not null |
 person_id        | uuid                     |           | not null |
 phone_number     | text                     |           | not null |
 created_at       | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
 modified_at      | timestamp with time zone |           | not null | CURRENT_TIMESTAMP
Indexes:
    "primary_contact_pkey" PRIMARY KEY, btree (id)
    "primary_contact_uniq_phone" UNIQUE, btree (location_id, phone_number)