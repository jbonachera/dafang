%::
	GOOS=linux GOARCH=mipsle go build -o out/motor ./cmd/motor/main.go && cat out/motor| pv  | ssh root@$@ "cat > /system/sdcard/bin/vx-motor"
