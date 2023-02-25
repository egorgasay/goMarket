package storage

import (
	"database/sql"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"gomarket/internal/schema"
	"log"
	"sync"
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
const updateBalance = `
UPDATE "Users"
SET "Balance" = "Balance" + $1
WHERE "Name" = $2
`
const changeOrerWithoutAccrual = `
UPDATE "Orders"
SET "Status" = $2
WHERE "UID" = $3
`
const checkBalance = `
SELECT 
    CASE
         WHEN "Balance" > $1 THEN 1
		 ELSE 2
    END
FROM "Users"
WHERE "Name" = $2
`
const drawBonuses = `
UPDATE "Users"
SET "Balance" = "Balance" - $1, 
    "Withdrawn" = "Withdrawn" + $1
WHERE "Name" = $2
`

func (s Storage) CreateUser(login, passwd string) error {
	prepare, err := s.DB.Prepare(createUser)
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
		return ErrUsernameConflict
	}

	return err
}

func (s Storage) CheckPassword(login, passwd string) error {
	prepare, err := s.DB.Prepare(validatePassword)
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
		return ErrWrongPassword
	}

	return err
}

func (s Storage) CheckID(username, id string) error {
	prepare, err := s.DB.Prepare(addOrder)
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
		prepareSecondQuery, err := s.DB.Prepare(getOwnerByID)
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
			return ErrCreatedByAnotherUser
		}

		return ErrCreatedByThisUser
	}

	return err
}

func (s Storage) GetOrders(username string) (Orders, error) {
	prepare, err := s.DB.Prepare(getOrders)
	if err != nil {
		return nil, err
	}

	rows, err := prepare.Query(username)
	if err != nil {
		return nil, err
	}

	orders := make(Orders, 0)

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
		return nil, ErrNoResult
	}

	return orders, nil
}

func (s Storage) GetBalance(username string) (schema.Balance, error) {
	prepare, err := s.DB.Prepare(getBalance)
	if err != nil {
		return schema.Balance{}, err
	}

	row := prepare.QueryRow(username)

	var balance schema.Balance
	return balance, row.Scan(&balance.Current, &balance.Withdrawn)
}

func (s Storage) UpdateOrder(username, id, status string, accrual float64) error {
	log.Println(status, accrual)
	if accrual == 0 {
		prepare, err := s.DB.Prepare(changeOrerWithoutAccrual)
		if err != nil {
			return err
		}

		_, err = prepare.Exec(status, id)
		if err != nil {
			return err
		}

		return nil
	}

	prepare, err := s.DB.Prepare(changeOrer)
	if err != nil {
		return err
	}

	_, err = prepare.Exec(accrual, status, id)
	if err != nil {
		return err
	}

	prepareBalance, err := s.DB.Prepare(updateBalance)
	if err != nil {
		return err
	}

	_, err = prepareBalance.Exec(accrual, username)
	return err
}

var usersBlock = make(map[string]*sync.Mutex)

func (s Storage) Withdraw(username string, amount float64, orderID string) error {
	usersBlock[username].Lock()
	defer usersBlock[username].Unlock()

	prepare, err := s.DB.Prepare(checkBalance)
	if err != nil {
		return err
	}

	row := prepare.QueryRow(amount, username) //, orderID)
	if row.Err() != nil {
		return err
	}

	var isEnoughMoney int
	err = row.Scan(&isEnoughMoney)
	if err != nil {
		return err
	}

	if isEnoughMoney == 2 {
		return ErrNotEnoughMoney
	}

	//if isEnoughMoney == 3 {
	//	return ErrWrongOrderID
	//}

	prepareDraw, err := s.DB.Prepare(drawBonuses)
	if err != nil {
		return err
	}

	_, err = prepareDraw.Exec(amount, username)
	if err != nil {
		return err
	}

	return nil
}
