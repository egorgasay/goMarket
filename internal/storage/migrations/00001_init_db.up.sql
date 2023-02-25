CREATE TABLE "Users" (
    "Name" VARCHAR(255) PRIMARY KEY,
    "Password" VARCHAR(255) NOT NULL,
    "Balance" DECIMAL NOT NULL,
    "Withdrawn" DECIMAL
);
CREATE TABLE "Orders" (
    "UID" VARCHAR(255) PRIMARY KEY,
    "Owner" VARCHAR(255) NOT NULL REFERENCES "Users"("Name"),
    "Date" TIMESTAMP NOT NULL,
    "Status" VARCHAR(255) CHECK (
        "Status" IN ('NEW', 'REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED')
    ),
    "Accrual" DECIMAL
);
CREATE TABLE Withdrawals (
    "Client" VARCHAR(255) NOT NULL REFERENCES "Users"("Name"),
    "UID" VARCHAR(255) NOT NULL REFERENCES "Orders"("UID"),
    "Sum" DECIMAL NOT NULL,
    "Date" TIMESTAMP NOT NULL
);