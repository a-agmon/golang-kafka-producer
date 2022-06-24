BIN_DIR = bin
PROTO_DIR = proto
PACKAGE = $(shell head -1 go.mod | awk '{print $$2}')
RM_F_CMD = rm -f
RM_RF_CMD = ${RM_F_CMD} -r

.PHONY: compile clean


compile:
	@echo "Executing $@"
	protoc -I${PROTO_DIR} --go_opt=module=${PACKAGE} --go_out=. --go-grpc_opt=module=${PACKAGE} --go-grpc_out=. ${PROTO_DIR}/*.proto
	go build -o ${BIN_DIR}/klient .


clean:
	${RM_F_CMD} ${PROTO_DIR}/*.pb.go

