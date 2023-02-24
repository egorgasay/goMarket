CREATE TABLE "Users" (
    "Name" VARCHAR(255) PRIMARY KEY NOT NULL,
    "Password" VARCHAR(255) NOT NULL,
    "Balance" MONEY NOT NULL
);
CREATE TABLE "Orders" (
    "Owner" VARCHAR(255) NOT NULL REFERENCES "Users"("Name"),
    "Date" TIMESTAMP NOT NULL,
    "UID" VARCHAR(255) NOT NULL,
    "Status" VARCHAR(255) CHECK (
        "Status" IN ('NEW', 'REGISTERED', 'INVALID', 'PROCESSING', 'PROCESSED')
    ),
    "Accrual" DECIMAL
);
