package authetication

import (
	"context"
	"testing"
)

type mockRepo struct {
}

func (m *mockRepo) CreateUser(ctx context.Context, email string, password string) (string, error) {
	c := service{}
	_, err := m.UserWithEmail(ctx, email)
	if err != nil {
		return "", ErrIdentityAlreadyExists
	}

	_, _ = c.CreateUser(context.Background(), email, password)
	return "", nil
}
func (m *mockRepo) UserWithEmail(ctx context.Context, email string) (string, error) {
	users := make(map[string]string)
	users["user@gmail.com"] = "12ddf"
	users["user1@gmail.com"] = "j332v"
	users["user2@gmail.com"] = "dv3rfesc"

	userid, ok := users[email]
	if !ok {

		return "", ErrIdentityAlreadyExists
	}
	return userid, nil
}
func (m *mockRepo) UserWithIdentity(ctx context.Context, email string, password string) (string, error) {

	users := make(map[string]string)
	users["user@gmail.com"] = "12ddf"
	users["user1@gmail.com"] = "j332v"
	users["user2@gmail.com"] = "dv3rfesc"

	userid, ok := users[email]
	if !ok {

		return "", ErrIdentityDoesNotExists
	}
	if password == "password" {
		return userid, nil
	}
	return "", ErrIdentityDoesNotExists
}
func (m *mockRepo) ResetPassword(ctx context.Context, email string) error {
	return nil

}

func Test_trimAndLower(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "email with forward space and cap",
			args: args{value: " JohnDoe"},
			want: "johndoe"},
		{name: "all caps with ending spaces",
			args: args{value: "JohnDOE    "},
			want: "johndoe"},
		{name: "just spaces",
			args: args{value: "   "},
			want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimAndLower(tt.args.value); got != tt.want {
				t.Errorf("trimAndLower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEmailValid(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "invalid email", args: args{email: "david@"}, want: false},
		{name: "proper email", args: args{email: "david@gmail.com"}, want: true},
		{name: "no email supplied", args: args{email: ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmailValid(tt.args.email); got != tt.want {
				t.Errorf("isEmailValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPasswordLengthValid(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "short password length", args: args{password: "12345"}, want: false},
		{name: "length of 6", args: args{password: "123456"}, want: true},
		{name: "length of 10", args: args{password: "1234567890"}, want: true},
		{name: "long password length", args: args{password: "12345123451234512345123451234512345123451234512345123451234512345123451234512345"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPasswordLengthValid(tt.args.password); got != tt.want {
				t.Errorf("isPasswordLengthValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TokenService(t *testing.T) {

	tests := []struct {
		name      string
		secretKey string
		arg       string
		want      string
	}{
		{name: "success", secretKey: "david secret key", arg: "david", want: "david"},
		{name: "success 2", secretKey: "david not so  secret key", arg: "david", want: "david"},
		{name: "success 3", secretKey: "david secret key", arg: "tomisin", want: "tomisin"},
		{name: "success 4", secretKey: "david not so secret key", arg: "tomisin", want: "tomisin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				secretKey: tt.secretKey,
			}
			token, _ := s.createToken(tt.arg)
			got, _ := s.VerifyToken(token)
			if got != tt.want {
				t.Errorf(" integration token service  got = %v, want %v", got, tt.want)
			}

		})

	}

}

func Test_service_Login(t *testing.T) {
	type fields struct {
		repo      Repository
		secretKey string
	}
	type args struct {
		email    string
		password string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "not registered email",
			fields:  fields{repo: &mockRepo{}, secretKey: "hello david"},
			args:    args{email: "david@gmail.com", password: "password"},
			want:    "",
			wantErr: true,
		},
		{name: "registered user 1",
			fields:  fields{secretKey: "hello david", repo: &mockRepo{}},
			args:    args{email: "user@gmail.com", password: "password"},
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTM4NjI5MzAsInN1YiI6IjEyZGRmIn0.V3OypgHfNIlcRA8HNgD2Q4ik4SAe5Php-_S5yX2MXJY",
			wantErr: false,
		},
		{name: "registered user 1 with false password",
			fields:  fields{secretKey: "hello david", repo: &mockRepo{}},
			args:    args{email: "user@gmail.com", password: "false"},
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTM4NjI5MzAsInN1YiI6IjEyZGRmIn0.V3OypgHfNIlcRA8HNgD2Q4ik4SAe5Php-_S5yX2MXJY",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo:      tt.fields.repo,
				secretKey: tt.fields.secretKey,
			}
			got, err := s.Login(context.Background(), tt.args.email, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			userid, ok := s.VerifyToken(got.Token)
			if !ok {
				t.Errorf("invalid token")
				return
			}
			useridtrue, _ := s.repo.UserWithEmail(context.Background(), tt.args.email)
			if userid != useridtrue {
				t.Errorf("userid does not match")
				return
			}
		})
	}
}
