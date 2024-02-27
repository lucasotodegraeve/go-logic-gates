alias w := watch

_default:
	just --list

watch:
	watchexec -e go -c clear go run	src/main.go
