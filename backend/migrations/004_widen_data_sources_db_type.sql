-- 既存インストール向け: data_sources.db_type の CHECK 制約を 4 種に広げる
-- init.sql の CREATE TABLE IF NOT EXISTS は既存テーブルを変更しないため、
-- 本 PR より前に作成された環境では別途この移行を適用する必要がある。
ALTER TABLE data_sources DROP CONSTRAINT IF EXISTS data_sources_db_type_check;
ALTER TABLE data_sources
    ADD CONSTRAINT data_sources_db_type_check
    CHECK (db_type IN ('postgresql', 'mysql', 'oracle', 'sqlserver'));
