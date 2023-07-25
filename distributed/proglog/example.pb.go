// example.proto 를 protobuf 컴파일러가 컴파일한다.
// 여러 언어로 컴파일 할 수 있으며, go 로 컴파일 하면 결과는 다음과 같다.
package twiter

type Tweet struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3"json:"message,omitempty"`
	// 내부 필드와 메서드는 생략
	// 인코딩/디코딩 메서드
	// payload가 적고, 직렬화가 빠르다.
}