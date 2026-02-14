package cmd

import (
	"github.com/urfave/cli/v3"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// Database represents the database command
var Database = &cli.Command{
	Name:  "database",
	Usage: "ezBookkeeping database maintenance",
	Commands: []*cli.Command{
		{
			Name:   "update",
			Usage:  "Update database structure",
			Action: bindAction(updateDatabaseStructure),
		},
	},
}

func updateDatabaseStructure(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] starting maintaining")

	err = updateAllDatabaseTablesStructure(c)

	if err != nil {
		log.CliErrorf(c, "[database.updateDatabaseStructure] update database table structure failed, because %s", err.Error())
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] all tables maintained successfully")
	return nil
}

func updateAllDatabaseTablesStructure(c *core.CliContext) error {
	var err error

	err = datastore.Container.UserStore.SyncStructs(new(models.User))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user table maintained successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactor))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor table maintained successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactorRecoveryCode))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor recovery code table maintained successfully")

	err = datastore.Container.TokenStore.SyncStructs(new(models.TokenRecord))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] token record table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Account))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] account table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Transaction))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction table maintained successfully")

	// Flatten all subcategories to top-level categories before syncing the table structure
	for i := 0; i < datastore.Container.UserDataStore.Count(); i++ {
		db := datastore.Container.UserDataStore.Get(i)
		_, flattenErr := db.NewSession(c).Exec("UPDATE transaction_category SET parent_category_id = 0 WHERE parent_category_id > 0")

		if flattenErr != nil {
			log.BootWarnf(c, "[database.updateAllDatabaseTablesStructure] flatten subcategories warning: %s", flattenErr.Error())
		} else {
			log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] subcategories flattened successfully")
		}
	}

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionCategory))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction category table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagGroup))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag group table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTag))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Counterparty))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] counterparty table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagIndex))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag index table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTemplate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction template table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionPictureInfo))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction picture table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionSplit))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction split table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserCustomExchangeRate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user custom exchange rate table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserApplicationCloudSetting))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user application cloud settings table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserExternalAuth))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user external auth table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.InsightsExplorer))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] insights explorer table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.CFO))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] cfo table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Location))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] location table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Asset))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] asset table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.InvestorDeal))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] investor_deal table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.InvestorPayment))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] investor_payment table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Budget))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] budget table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Obligation))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] obligation table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TaxRecord))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] tax_record table maintained successfully")

	return nil
}
