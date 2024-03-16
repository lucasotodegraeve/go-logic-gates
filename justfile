alias w := watch

_default:
	just --list

watch target:
	watchexec -c clear -e go -r go run {{target}}
