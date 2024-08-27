pwd = $$(pwd)

build = go build
build_container = docker build
build_flags_windows = GOOS=windows GOARCH=amd64

build_dir = $(pwd)/build
build_bin_dir = $(build_dir)/bin
asm_container_name = chariot-asm

all: cli asm bas

cli: .build_bin_dir
	@$(build) -o $(build_bin_dir)/chariot ./cmd/chariot

asm:
	@$(build_container) -f $(build_dir)/asm/Dockerfile -t $(asm_container_name) $(pwd)

bas: .build_bin_dir
	@for f in $$(find ./cmd/bas -type f -name "*.go"); do $(build_flags_windows) $(build) -o $(build_bin_dir)/$$(echo $$f | cut -d'/' -f4 | sed "s/\.go//") $$f; done

clean:
	@rm -rf $(build_bin_dir)
	@docker rmi -f $(asm_container_name) 2>/dev/null

.build_bin_dir:
	@mkdir -p $(build_bin_dir)
