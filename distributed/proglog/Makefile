CONFIG_PATH = ${HOMEPATH}\.proglog
.PHONY: init
init:
	mkdir ${CONFIG_PATH}

.PHONY: gencert
gencert:
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server
	
	xcopy .\ca-key.pem ${CONFIG_PATH}
	xcopy .\ca.csr ${CONFIG_PATH}
	xcopy .\ca.pem ${CONFIG_PATH}
	xcopy .\server-key.pem ${CONFIG_PATH}
	xcopy .\server.csr ${CONFIG_PATH}
	xcopy .\server.pem ${CONFIG_PATH}

	del .\ca-key.pem
	del .\ca.csr
	del .\ca.pem 
	del .\server-key.pem 
	del .\server.csr
	del .\server.pem 

.PHONY: compile
compile:
	protoc api/v1/*.proto --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=.

.PHONY: test
test:
	go test -race -v ./...