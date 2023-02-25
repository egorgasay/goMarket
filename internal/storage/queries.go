package storage

const createUser = `INSERT INTO "Users" VALUES ($1, $2, 0.0, 0.0)`
const validatePassword = `
SELECT 1 FROM "Users" WHERE "Name" = $1 AND "Password" = $2
`
const addOrder = `
INSERT INTO "Orders" VALUES ($2, $1, now()::timestamp, 'NEW', 0)
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
const stageDraw = `
INSERT INTO Withdrawals VALUES ($1, $2, $3, now()::timestamp)
`
const getWithdrawals = `
SELECT "UID", "Sum", "Date" FROM Withdrawals WHERE "Client" = $1
`
