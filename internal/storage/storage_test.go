package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/egorgasay/dockerdb"
	"gomarket/internal/schema"
	"log"
	"os"
	"reflect"
	"testing"
)

var TestDB Storage

func TestMain(m *testing.M) {
	// Write code here to run before tests
	ctx := context.TODO()
	cfg := dockerdb.CustomDB{
		DB: dockerdb.DB{
			Name:     "vdb_te45",
			User:     "admin",
			Password: "admin",
		},
		Port: "12545",
		Vendor: dockerdb.Vendor{
			Name:  dockerdb.Postgres,
			Image: "postgres", // TODO: add dockerdb.Postgres15 as image into dockerdb package
		},
	}
	vdb, err := dockerdb.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	TestDB = New(vdb.DB, "file://migrations").(Storage)

	queries := []string{
		"DROP SCHEMA public CASCADE;",
		"CREATE SCHEMA public;",
		"GRANT ALL ON SCHEMA public TO public;",
		"COMMENT ON SCHEMA public IS 'standard public schema';",
	}

	tx, err := TestDB.DB.Begin()
	if err != nil {
		log.Fatal(err)
	}

	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	TestDB = New(vdb.DB, "file://migrations").(Storage)
	// Run tests
	exitVal := m.Run()

	// Write code here to run after tests
	// Exit with exit value from tests
	os.Exit(exitVal)
}

func TestStorage_CreateUser(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin", "admin"},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "wrong",
			args: args{"admin", "1234"},
			err:  Err{want: true, Error: ErrUsernameConflict},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestDB.CreateUser(tt.args.username, tt.args.password); (err != nil) != tt.err.want {
				t.Errorf("CreateUser() error = %v, \nwantErr %v", err, tt.err.want)
			} else if tt.err.want && !errors.Is(err, tt.err.Error) {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.err.Error)
			}
		})
	}
}

func TestStorage_CheckPassword(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name string
		args args
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin", "admin"},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "wrong",
			args: args{"admin", "1234"},
			err:  Err{want: true, Error: ErrWrongPassword},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestDB.CheckPassword(tt.args.username, tt.args.password); (err != nil) != tt.err.want {
				t.Errorf("CheckPassword() error = %v, \nwantErr %v", err, tt.err.want)
			} else if tt.err.want && !errors.Is(err, tt.err.Error) {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.err.Error)
			}
		})
	}
}

func TestStorage_CheckID(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
		id       string
	}
	tests := []struct {
		name string
		args args
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin", "1234"},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "ErrCreatedByAnotherUser",
			args: args{"admin2", "1234"},
			err:  Err{want: true, Error: ErrCreatedByAnotherUser},
		},
		{
			name: "ErrCreatedByThisUser",
			args: args{"admin", "1234"},
			err:  Err{want: true, Error: ErrCreatedByThisUser},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestDB.CheckID(tt.args.username, tt.args.id); (err != nil) != tt.err.want {
				t.Errorf("CheckID() error = %v, \nwantErr %v", err, tt.err.want)
			} else if tt.err.want && !errors.Is(err, tt.err.Error) {
				t.Errorf("CheckID() error = %v, wantErr %v", err, tt.err.Error)
			}
		})
	}
}

func TestStorage_GetBalance(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
	}
	tests := []struct {
		name string
		args args
		want schema.Balance
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin"},
			want: schema.Balance{Current: 0, Withdrawn: 0},
			err:  Err{want: false, Error: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TestDB.GetBalance(tt.args.username)
			if (err != nil) != tt.err.want {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.err.want)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_GetOrders(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
	}
	tests := []struct {
		name string
		args args
		want Orders
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin"},
			want: Orders{schema.UserOrder{Number: "1234", Status: "NEW"}},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "no res",
			args: args{"admin1567"},
			want: Orders{},
			err:  Err{want: true, Error: ErrNoResult},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TestDB.GetOrders(tt.args.username)
			if (err != nil) != tt.err.want {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.err.want)
				return
			} else if !errors.Is(err, tt.err.Error) {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.err.Error)
				return
			}
			if !tt.err.want && (tt.want[0].Number != "1234" || tt.want[0].Status != "NEW") {
				t.Errorf("GetOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_Withdraw(t *testing.T) {
	_, err := TestDB.DB.Exec(`UPDATE "Users" SET "Balance" = 100.00 `)
	if err != nil {
		t.Fatal(err)
	}
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
		amount   float64
		orderID  string
	}
	tests := []struct {
		name string
		args args
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin", 1, "1234"},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "not enough money",
			args: args{"admin", 1234, "1234"},
			err:  Err{want: true, Error: ErrNotEnoughMoney},
		},
		{
			name: "bad user",
			args: args{"admin1", 1234, "1234"},
			err:  Err{want: true, Error: sql.ErrNoRows},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestDB.Withdraw(tt.args.username, tt.args.amount, tt.args.orderID); (err != nil) != tt.err.want {
				t.Errorf("Withdraw() error = %v, \nwantErr %v", err, tt.err.want)
			} else if tt.err.want && !errors.Is(err, tt.err.Error) {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.err.Error)
			}
		})
	}
}

func TestStorage_GetWithdrawals(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
	}
	tests := []struct {
		name string
		args args
		want []schema.Withdrawn
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin"},
			want: []schema.Withdrawn{schema.Withdrawn{Order: "1234", Sum: 1}},
			err:  Err{want: false, Error: nil},
		},
		{
			name: "no res",
			args: args{"admin1567"},
			want: []schema.Withdrawn{schema.Withdrawn{}},
			err:  Err{want: true, Error: ErrNoWithdrawals},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TestDB.GetWithdrawals(tt.args.username)
			if (err != nil) != tt.err.want {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.err.want)
				return
			} else if !errors.Is(err, tt.err.Error) {
				t.Errorf("GetOrders() error = %v, wantErr %v", err, tt.err.Error)
				return
			}
			if !tt.err.want && (tt.want[0].Order != "1234" || tt.want[0].Sum != 1) {
				t.Errorf("GetOrders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_UpdateOrder(t *testing.T) {
	type Err struct {
		want  bool
		Error error
	}
	type args struct {
		username string
		id       string
		status   string
		accrual  float64
	}
	tests := []struct {
		name string
		args args
		err  Err
	}{
		{
			name: "ok",
			args: args{"admin", "1234", "INVALID", 500},
			err:  Err{want: false, Error: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestDB.UpdateOrder(tt.args.username, tt.args.id, tt.args.status, tt.args.accrual); (err != nil) != tt.err.want {
				t.Errorf("UpdateOrder() error = %v, \nwantErr %v", err, tt.err.want)
			} else if tt.err.want && !errors.Is(err, tt.err.Error) {
				t.Errorf("UpdateOrder() error = %v, wantErr %v", err, tt.err.Error)
			}
		})
	}
}
