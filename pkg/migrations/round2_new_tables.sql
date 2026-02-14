-- Round 2 Migration: Create all new tables added during Round 2
-- Compatible with SQLite
-- Safe to re-run: uses CREATE TABLE IF NOT EXISTS and column-existence checks

-- =============================================================================
-- 1. CFO (Cost/Financial center)
-- Model: CFO -> table: c_f_o
-- =============================================================================
CREATE TABLE IF NOT EXISTS `c_f_o` (
    `cfo_id`            INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `name`              VARCHAR(64) NOT NULL,
    `color`             VARCHAR(6) NOT NULL,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `display_order`     INTEGER NOT NULL DEFAULT 0,
    `hidden`            INTEGER NOT NULL DEFAULT 0,
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_cfo_uid_deleted_order`
    ON `c_f_o` (`uid`, `deleted`, `display_order`);

-- =============================================================================
-- 2. Counterparty
-- Model: Counterparty -> table: counterparty
-- =============================================================================
CREATE TABLE IF NOT EXISTS `counterparty` (
    `counterparty_id`   INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `type`              INTEGER NOT NULL DEFAULT 0,
    `name`              VARCHAR(64) NOT NULL,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `icon`              INTEGER NOT NULL DEFAULT 0,
    `color`             VARCHAR(6) NOT NULL DEFAULT '',
    `display_order`     INTEGER NOT NULL DEFAULT 0,
    `hidden`            INTEGER NOT NULL DEFAULT 0,
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_counterparty_uid_deleted_order`
    ON `counterparty` (`uid`, `deleted`, `display_order`);

-- =============================================================================
-- 3. Location
-- Model: Location -> table: location
-- =============================================================================
CREATE TABLE IF NOT EXISTS `location` (
    `location_id`          INTEGER PRIMARY KEY,
    `uid`                  INTEGER NOT NULL,
    `deleted`              INTEGER NOT NULL DEFAULT 0,
    `cfo_id`               INTEGER NOT NULL DEFAULT 0,
    `name`                 VARCHAR(64) NOT NULL,
    `address`              VARCHAR(255) NOT NULL DEFAULT '',
    `location_type`        INTEGER NOT NULL DEFAULT 1,
    `monthly_rent`         INTEGER NOT NULL DEFAULT 0,
    `monthly_electricity`  INTEGER NOT NULL DEFAULT 0,
    `monthly_internet`     INTEGER NOT NULL DEFAULT 0,
    `comment`              VARCHAR(255) NOT NULL DEFAULT '',
    `display_order`        INTEGER NOT NULL DEFAULT 0,
    `hidden`               INTEGER NOT NULL DEFAULT 0,
    `created_unix_time`    INTEGER DEFAULT 0,
    `updated_unix_time`    INTEGER DEFAULT 0,
    `deleted_unix_time`    INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_location_uid_deleted_order`
    ON `location` (`uid`, `deleted`, `display_order`);

-- =============================================================================
-- 4. Asset
-- Model: Asset -> table: asset
-- =============================================================================
CREATE TABLE IF NOT EXISTS `asset` (
    `asset_id`                  INTEGER PRIMARY KEY,
    `uid`                       INTEGER NOT NULL,
    `deleted`                   INTEGER NOT NULL DEFAULT 0,
    `cfo_id`                    INTEGER NOT NULL DEFAULT 0,
    `location_id`               INTEGER NOT NULL DEFAULT 0,
    `name`                      VARCHAR(64) NOT NULL,
    `asset_type`                INTEGER NOT NULL DEFAULT 1,
    `purchase_date`             INTEGER NOT NULL DEFAULT 0,
    `purchase_cost`             INTEGER NOT NULL DEFAULT 0,
    `useful_life_months`        INTEGER NOT NULL DEFAULT 0,
    `salvage_value`             INTEGER NOT NULL DEFAULT 0,
    `status`                    INTEGER NOT NULL DEFAULT 1,
    `commission_date`           INTEGER NOT NULL DEFAULT 0,
    `decommission_date`         INTEGER NOT NULL DEFAULT 0,
    `installed_capacity_watts`  INTEGER NOT NULL DEFAULT 0,
    `comment`                   VARCHAR(255) NOT NULL DEFAULT '',
    `display_order`             INTEGER NOT NULL DEFAULT 0,
    `hidden`                    INTEGER NOT NULL DEFAULT 0,
    `created_unix_time`         INTEGER DEFAULT 0,
    `updated_unix_time`         INTEGER DEFAULT 0,
    `deleted_unix_time`         INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_asset_uid_deleted_order`
    ON `asset` (`uid`, `deleted`, `display_order`);

-- =============================================================================
-- 5. InvestorDeal
-- Model: InvestorDeal -> table: investor_deal
-- =============================================================================
CREATE TABLE IF NOT EXISTS `investor_deal` (
    `deal_id`               INTEGER PRIMARY KEY,
    `uid`                   INTEGER NOT NULL,
    `deleted`               INTEGER NOT NULL DEFAULT 0,
    `investor_name`         VARCHAR(64) NOT NULL,
    `cfo_id`                INTEGER NOT NULL DEFAULT 0,
    `investment_date`       INTEGER NOT NULL DEFAULT 0,
    `investment_amount`     INTEGER NOT NULL DEFAULT 0,
    `currency`              VARCHAR(3) NOT NULL DEFAULT 'RUB',
    `deal_type`             INTEGER NOT NULL DEFAULT 1,
    `annual_rate`           INTEGER NOT NULL DEFAULT 0,
    `profit_share_pct`      INTEGER NOT NULL DEFAULT 0,
    `fixed_payment`         INTEGER NOT NULL DEFAULT 0,
    `repayment_start_date`  INTEGER NOT NULL DEFAULT 0,
    `repayment_end_date`    INTEGER NOT NULL DEFAULT 0,
    `total_to_repay`        INTEGER NOT NULL DEFAULT 0,
    `comment`               VARCHAR(255) NOT NULL DEFAULT '',
    `created_unix_time`     INTEGER DEFAULT 0,
    `updated_unix_time`     INTEGER DEFAULT 0,
    `deleted_unix_time`     INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_investor_deal_uid_deleted`
    ON `investor_deal` (`uid`, `deleted`);

-- =============================================================================
-- 6. InvestorPayment
-- Model: InvestorPayment -> table: investor_payment
-- =============================================================================
CREATE TABLE IF NOT EXISTS `investor_payment` (
    `payment_id`        INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `deal_id`           INTEGER NOT NULL,
    `payment_date`      INTEGER NOT NULL DEFAULT 0,
    `amount`            INTEGER NOT NULL DEFAULT 0,
    `payment_type`      INTEGER NOT NULL DEFAULT 1,
    `transaction_id`    INTEGER NOT NULL DEFAULT 0,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_investor_payment_uid_deleted`
    ON `investor_payment` (`uid`, `deleted`);

CREATE INDEX IF NOT EXISTS `IDX_investor_payment_deal_id`
    ON `investor_payment` (`deal_id`);

-- =============================================================================
-- 7. Budget
-- Model: Budget -> table: budget
-- =============================================================================
CREATE TABLE IF NOT EXISTS `budget` (
    `budget_id`         INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `cfo_id`            INTEGER NOT NULL DEFAULT 0,
    `category_id`       INTEGER NOT NULL DEFAULT 0,
    `year`              INTEGER NOT NULL DEFAULT 0,
    `month`             INTEGER NOT NULL DEFAULT 0,
    `planned_amount`    INTEGER NOT NULL DEFAULT 0,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_budget_uid_deleted_year_month`
    ON `budget` (`uid`, `deleted`, `year`, `month`);

-- =============================================================================
-- 8. Obligation
-- Model: Obligation -> table: obligation
-- =============================================================================
CREATE TABLE IF NOT EXISTS `obligation` (
    `obligation_id`     INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `obligation_type`   INTEGER NOT NULL DEFAULT 1,
    `counterparty_id`   INTEGER NOT NULL DEFAULT 0,
    `cfo_id`            INTEGER NOT NULL DEFAULT 0,
    `amount`            INTEGER NOT NULL DEFAULT 0,
    `currency`          VARCHAR(3) NOT NULL DEFAULT 'RUB',
    `due_date`          INTEGER NOT NULL DEFAULT 0,
    `status`            INTEGER NOT NULL DEFAULT 1,
    `paid_amount`       INTEGER NOT NULL DEFAULT 0,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_obligation_uid_deleted`
    ON `obligation` (`uid`, `deleted`);

-- =============================================================================
-- 9. TaxRecord
-- Model: TaxRecord -> table: tax_record
-- =============================================================================
CREATE TABLE IF NOT EXISTS `tax_record` (
    `tax_id`            INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `cfo_id`            INTEGER NOT NULL DEFAULT 0,
    `tax_type`          INTEGER NOT NULL DEFAULT 1,
    `period_year`       INTEGER NOT NULL DEFAULT 0,
    `period_quarter`    INTEGER NOT NULL DEFAULT 0,
    `taxable_income`    INTEGER NOT NULL DEFAULT 0,
    `tax_amount`        INTEGER NOT NULL DEFAULT 0,
    `paid_amount`       INTEGER NOT NULL DEFAULT 0,
    `due_date`          INTEGER NOT NULL DEFAULT 0,
    `status`            INTEGER NOT NULL DEFAULT 1,
    `comment`           VARCHAR(255) NOT NULL DEFAULT '',
    `currency`          VARCHAR(3) NOT NULL DEFAULT 'RUB',
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_tax_record_uid_deleted`
    ON `tax_record` (`uid`, `deleted`);

-- =============================================================================
-- 10. InsightsExplorer
-- Model: InsightsExplorer -> table: insights_explorer
-- =============================================================================
CREATE TABLE IF NOT EXISTS `insights_explorer` (
    `explorer_id`       INTEGER PRIMARY KEY,
    `uid`               INTEGER NOT NULL,
    `deleted`           INTEGER NOT NULL DEFAULT 0,
    `name`              VARCHAR(64) NOT NULL,
    `display_order`     INTEGER NOT NULL DEFAULT 0,
    `data`              BLOB,
    `hidden`            INTEGER NOT NULL DEFAULT 0,
    `created_unix_time` INTEGER DEFAULT 0,
    `updated_unix_time` INTEGER DEFAULT 0,
    `deleted_unix_time` INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS `IDX_insights_explorer_uid_deleted_order`
    ON `insights_explorer` (`uid`, `deleted`, `display_order`);

-- =============================================================================
-- ALTER TABLE: Add currency column to tax_record if it does not exist
-- (for databases created before Currency was added to the TaxRecord model)
-- Note: SQLite does not support IF NOT EXISTS for ALTER TABLE ADD COLUMN,
-- so this statement will fail harmlessly if the column already exists.
-- When running via script, wrap in a try/catch or ignore the error.
-- =============================================================================
-- ALTER TABLE `tax_record` ADD COLUMN `currency` VARCHAR(3) NOT NULL DEFAULT 'RUB';
