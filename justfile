alias w := watch

_default:
	just --list

watch:
	watchexec -c clear -e go -r go run main.go
