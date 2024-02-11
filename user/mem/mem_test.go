package mem

import (
	"context"
	"reflect"
	"testing"

	"github.com/DanCreative/veracode-admin-plus/user"
)

func setup() *UserLocalMemRepository {
	cart := map[string]user.User{
		"0":  {UserId: "0", EmailAddress: "0@example.com"},
		"1":  {UserId: "1", EmailAddress: "1@example.com"},
		"2":  {UserId: "2", EmailAddress: "2@example.com"},
		"3":  {UserId: "3", EmailAddress: "3@example.com"},
		"4":  {UserId: "4", EmailAddress: "4@example.com"},
		"5":  {UserId: "5", EmailAddress: "5@example.com"},
		"6":  {UserId: "6", EmailAddress: "6@example.com"},
		"7":  {UserId: "7", EmailAddress: "7@example.com"},
		"8":  {UserId: "8", EmailAddress: "8@example.com"},
		"9":  {UserId: "9", EmailAddress: "9@example.com"},
		"10": {UserId: "10", EmailAddress: "10@example.com"},
		"11": {UserId: "11", EmailAddress: "11@example.com"},
	}

	repo := UserLocalMemRepository{
		userCart: cart,
	}

	return &repo
}

func TestUserLocalMemRepository_GetCartUsers(t *testing.T) {
	type args struct {
		ctx     context.Context
		options user.SearchUserOptions
	}
	tests := []struct {
		name    string
		ulr     *UserLocalMemRepository
		args    args
		want    []user.User
		want1   user.PageMeta
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "cart_page_1",
			ulr:  setup(),
			want1: user.PageMeta{
				PageNumber:    0,
				Size:          10,
				TotalElements: 12,
				TotalPages:    2,
				FirstParams:   "cart=Yes&page=0&size=10",
				LastParams:    "cart=Yes&page=1&size=10",
				SelfParams:    "cart=Yes&page=0&size=10",
				NextParams:    "cart=Yes&page=1&size=10",
			},
			args: args{ctx: context.Background(), options: user.SearchUserOptions{
				Page: 0,
				Size: 10,
				Cart: "Yes",
			}},
			want: []user.User{
				{UserId: "0", EmailAddress: "0@example.com"},
				{UserId: "1", EmailAddress: "1@example.com"},
				{UserId: "2", EmailAddress: "2@example.com"},
				{UserId: "3", EmailAddress: "3@example.com"},
				{UserId: "4", EmailAddress: "4@example.com"},
				{UserId: "5", EmailAddress: "5@example.com"},
				{UserId: "6", EmailAddress: "6@example.com"},
				{UserId: "7", EmailAddress: "7@example.com"},
				{UserId: "8", EmailAddress: "8@example.com"},
				{UserId: "9", EmailAddress: "9@example.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.ulr.GetCartUsers(tt.args.ctx, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserLocalMemRepository.GetCartUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserLocalMemRepository.GetCartUsers() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("UserLocalMemRepository.GetCartUsers() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
