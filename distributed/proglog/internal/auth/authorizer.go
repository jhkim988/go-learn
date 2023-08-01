/*
인증을 한 후에 특정 행위에 대한 권한 확인을 한다.
ACL: Access Control List, 접근제어목록. "subject 는 object 에 action 할 권한이 있다"는 식으로 저장된다.
ACL 은 테이블일 뿐이며, Map 이나 csv, 관계형 데이터베이스 등으로 전환할 수 있다.
Casbin 이라는 라이브러리를 이용
*/
package auth

import (
	"fmt"

	"github.com/casbin/casbin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func New(model, policy string) *Authorizer {
	enforcer := casbin.NewEnforcer(model, policy)
	return &Authorizer {
		enforcer: enforcer,
	}
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	/* 권한이 없다면 error 리턴한다. */
	if !a.enforcer.Enforce(subject, object, action) {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}

