APP := build/MorphoWheel.app

BIN_DIR  := $(HOME)/.local/bin
APP_DIR  := $(HOME)/Applications
LABEL    := local.morphowheel.mac
AGENT    := $(HOME)/Library/LaunchAgents/$(LABEL).plist
INSTALLED := $(APP_DIR)/MorphoWheel.app

.PHONY: build app test run clean install uninstall autostart-on autostart-off
build:
	go build -o MorphoWheel .

app: build
	rm -rf $(APP) && mkdir -p $(APP)/Contents/MacOS
	cp MorphoWheel $(APP)/Contents/MacOS/MorphoWheel
	cp Info.plist $(APP)/Contents/Info.plist
	codesign --force --sign - $(APP)

test:
	go vet ./... && go test ./...

run: app
	open $(APP)

clean:
	rm -rf build MorphoWheel

# Copies the app to ~/Applications and links the binary into ~/.local/bin.
install: app
	mkdir -p $(APP_DIR) $(BIN_DIR)
	rm -rf $(INSTALLED) && cp -R $(APP) $(INSTALLED)
	ln -sf $(INSTALLED)/Contents/MacOS/MorphoWheel $(BIN_DIR)/morphowheel
	@echo "installed: $(INSTALLED), command: morphowheel"

uninstall: autostart-off
	rm -rf $(INSTALLED) $(BIN_DIR)/morphowheel

# Start at login through a LaunchAgent.
autostart-on: install
	mkdir -p $(dir $(AGENT))
	printf '%s\n' \
	  '<?xml version="1.0" encoding="UTF-8"?>' \
	  '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
	  '<plist version="1.0"><dict>' \
	  '<key>Label</key><string>$(LABEL)</string>' \
	  '<key>ProgramArguments</key><array><string>$(INSTALLED)/Contents/MacOS/MorphoWheel</string></array>' \
	  '<key>RunAtLoad</key><true/>' \
	  '<key>KeepAlive</key><dict><key>SuccessfulExit</key><false/></dict>' \
	  '<key>ProcessType</key><string>Interactive</string>' \
	  '</dict></plist>' > $(AGENT)
	-launchctl bootout gui/$$(id -u)/$(LABEL) 2>/dev/null
	launchctl bootstrap gui/$$(id -u) $(AGENT)

autostart-off:
	-launchctl bootout gui/$$(id -u)/$(LABEL) 2>/dev/null
	rm -f $(AGENT)
