prepare:
	curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v3.0.1/aqua-installer | bash
	@echo ''
	@echo 'Add $${AQUA_ROOT_DIR}/bin to the environment variable PATH.'
	@echo 'export PATH="$${AQUA_ROOT_DIR:-$${XDG_DATA_HOME:-$$HOME/.local/share}/aquaproj-aqua}/bin:$$PATH"'

proto:
	cd plugin/internal/proto; \
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative tflint.proto
