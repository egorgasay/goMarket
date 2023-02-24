package postgres

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"gomarket/internal/schema"
	"gomarket/internal/storage/service"
	"log"
)

const createUser = `INSERT INTO "Users" VALUES ($1, $2, 0.0, 0.0)`
const validatePassword = `
SELECT 1 FROM "Users" WHERE "Name" = $1 AND "Password" = $2
`
const addOrder = `
INSERT INTO "Orders" VALUES ($1, now()::timestamp, $2, 'NEW', 0)
`
const getOwnerByID = `
SELECT "Owner" FROM "Orders" WHERE "UID" = $1
`
const getOrders = `
SELECT "UID", "Status", "Accrual", "Date" FROM "Orders" WHERE "Owner" = $1
`
const getBalance = `
SELECT "Balance", "Withdrawn" FROM "Users" WHERE "Name" = $1
`
const changeOrer = `
UPDATE "Orders"
SET "Accrual" = $1,
    "Status" = $2
WHERE "UID" = $3
`

type Postgres struct {
	DB *sql.DB
}

func New(db *sql.DB, pathToMigrations string) service.IRealStorage {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	m, err := migrate.NewWithDatabaseInstance(
		pathToMigrations,
		"gomarket", driver)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	err = m.Up()
	if err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}

	return Postgres{DB: db}
}

func (p Postgres) CreateUser(login, passwd string) error {
	prepare, err := p.DB.Prepare(createUser)
	if err != nil {
		return err
	}

	_, err = prepare.Exec(login, passwd)
	if err == nil {
		return nil
	}

	e, ok := err.(*pq.Error)
	if !ok {
		log.Println("shouldn't be this ", err)
		return err
	}

	if e.Code == pgerrcode.UniqueViolation {
		return service.ErrUsernameConflict
	}

	return err
}

func (p Postgres) CheckPassword(login, passwd string) error {
	prepare, err := p.DB.Prepare(validatePassword)
	if err != nil {
		return err
	}

	row := prepare.QueryRow(login, passwd)
	if row.Err() != nil {
		return err
	}

	var isValidPassword bool
	err = row.Scan(&isValidPassword)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrWrongPassword
	}

	return err
}

func (p Postgres) CheckID(username, id string) error {
	prepare, err := p.DB.Prepare(addOrder)
	if err != nil {
		return err
	}

	_, err = prepare.Exec(username, id)
	if err == nil {
		return nil
	}

	e, ok := err.(*pq.Error)
	if !ok {
		log.Println("shouldn't be this ", err)
		return err
	}

	if e.Code == pgerrcode.UniqueViolation {
		prepareSecondQuery, err := p.DB.Prepare(getOwnerByID)
		if err != nil {
			return err
		}

		var owner string
		row := prepareSecondQuery.QueryRow(id)

		err = row.Scan(&owner)
		if err != nil {
			return err
		}

		if owner != username {
			return service.ErrCreatedByAnotherUser
		}

		return service.ErrCreatedByThisUser
	}

	return err
}

func (p Postgres) GetOrders(username string) (service.Orders, error) {
	prepare, err := p.DB.Prepare(getOrders)
	if err != nil {
		return nil, err
	}

	rows, err := prepare.Query(username)
	if err != nil {
		return nil, err
	}

	orders := make(service.Orders, 0)

	for rows.Next() {
		order := schema.UserOrder{}
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, service.ErrNoResult
	}

	return orders, nil
}

func (p Postgres) GetBalance(username string) (schema.Balance, error) {
	prepare, err := p.DB.Prepare(getBalance)
	if err != nil {
		return schema.Balance{}, err
	}

	row := prepare.QueryRow(username)

	var balance schema.Balance
	return balance, row.Scan(&balance.Current, &balance.Withdrawn)
}

func (p Postgres) UpdateOrder(id, status string, accrual int) error {
	prepare, err := p.DB.Prepare(changeOrer)
	if err != nil {
		return err
	}

	_, err = prepare.Exec(accrual, status, id)
	return err
}
