#!/bin/sh
# Skopos supervisor: exit 0 stops, 75 restarts immediately, other codes back off.
# Forward termination signals and wait for the child to finish graceful shutdown.
# Download verification and atomic replacement are handled by selfupdate.Bootstrap.
set -u
LANG_SEL="${SKOPOS_LANGUAGE:-${SKOPOS_LANGUAGE:-en}}"
case "$LANG_SEL" in ko|ko_*|ko-*|KO) LANG_SEL=ko ;; *) LANG_SEL=en ;; esac
msg(){ if [ "$LANG_SEL" = ko ]; then printf '%s' "$2"; else printf '%s' "$1"; fi; }


cd "$(dirname "$0")" || exit 1

BIN=./skopos
[ -x "$BIN" ] || { echo "[Skopos] $(msg "Executable not found: $BIN" "실행 파일을 찾을 수 없습니다: $BIN")" >&2; exit 1; }

RESTART_CODE=75
MAX_DELAY=60

child=0
stopping=0

forward() {
	stopping=1
	if [ "$child" -ne 0 ]; then
		kill -TERM "$child" 2>/dev/null || true
	fi
}
trap forward INT TERM

# Ports this process is about to bind. An empty proxy address disables that listener.
listen_ports() {
	mgmt_addr=127.0.0.1:8787
	proxy_addr=127.0.0.1:8788
	flag_name=
	for arg in "$@"; do
		if [ -n "$flag_name" ]; then
			case "$flag_name" in
				addr) mgmt_addr=$arg ;;
				proxy) proxy_addr=$arg ;;
			esac
			flag_name=
			continue
		fi
		case "$arg" in
			-addr|--addr) flag_name=addr ;;
			-proxy|--proxy) flag_name=proxy ;;
			-addr=*|--addr=*) mgmt_addr=${arg#*=} ;;
			-proxy=*|--proxy=*) proxy_addr=${arg#*=} ;;
		esac
	done
	printf '%s\n' "$mgmt_addr" | awk -F: 'NF { print $NF }'
	if [ -n "$proxy_addr" ]; then
		printf '%s\n' "$proxy_addr" | awk -F: 'NF { print $NF }'
	fi
}

listen_pids() {
	if ! command -v lsof >/dev/null 2>&1; then
		if [ "${lsof_warned:-0}" -eq 0 ]; then
			echo "[Skopos] $(msg "lsof is required to close a busy listen port" "사용 중인 포트를 닫으려면 lsof가 필요합니다")" >&2
			lsof_warned=1
		fi
		return 0
	fi
	lsof -nP -iTCP:"$1" -sTCP:LISTEN -t 2>/dev/null | awk 'NF && !seen[$1]++ { print $1 }'
}

stop_listener() {
	port_pid=$1
	case "$port_pid" in
		''|*[!0-9]*|0|1|$$) return ;;
	esac
	if [ "$child" -ne 0 ] && [ "$port_pid" -eq "$child" ]; then
		return
	fi
	port_ppid=$(ps -o ppid= -p "$port_pid" 2>/dev/null | tr -d '[:space:]')
	case "$port_ppid" in
		''|*[!0-9]*|0|1|$$) ;;
		*)
			port_pcmd=$(ps -o command= -p "$port_ppid" 2>/dev/null || true)
			case "$port_pcmd" in
				*[\ /]start.sh|*[\ /]start.sh\ *)
					echo "[Skopos] $(msg "Stopping supervisor $port_ppid so it does not relaunch" "다시 실행하지 않도록 감독 프로세스 $port_ppid 중지")" >&2
					kill -TERM "$port_ppid" 2>/dev/null || true
					;;
			esac
			;;
	esac
	port_cmd=$(ps -o command= -p "$port_pid" 2>/dev/null || true)
	echo "[Skopos] $(msg "Port $2 is open (pid $port_pid: $port_cmd); closing it" "포트 $2 이(가) 열려 있습니다 (pid $port_pid: $port_cmd); 닫습니다")" >&2
	kill -TERM "$port_pid" 2>/dev/null || true
}

release_port() {
	rel_port=$1
	case "$rel_port" in
		''|*[!0-9]*) return ;;
	esac
	rel_pids=$(listen_pids "$rel_port")
	[ -n "$rel_pids" ] || return 0
	for port_pid in $rel_pids; do
		stop_listener "$port_pid" "$rel_port"
	done
	rel_n=0
	while [ "$rel_n" -lt 5 ]; do
		rel_left=$(listen_pids "$rel_port")
		[ -z "$rel_left" ] && return 0
		sleep 1
		rel_n=$((rel_n + 1))
	done
	for port_pid in $rel_left; do
		case "$port_pid" in
			''|*[!0-9]*|0|1|$$) continue ;;
		esac
		echo "[Skopos] $(msg "Port $rel_port is still open (pid $port_pid); forcing it closed" "포트 $rel_port 이(가) 여전히 열려 있습니다 (pid $port_pid); 강제 종료")" >&2
		kill -KILL "$port_pid" 2>/dev/null || true
	done
	sleep 1
}

release_listen_ports() {
	rel_ports=$(listen_ports "$@")
	for rel_port in $rel_ports; do
		release_port "$rel_port"
	done
}

delay=1
while :; do
	release_listen_ports "$@"
	"$BIN" "$@" &
	child=$!

	wait "$child"
	code=$?
	if [ "$code" -gt 128 ]; then
		wait "$child"
		code=$?
	fi
	child=0

	if [ "$stopping" -eq 1 ]; then
		echo "[Skopos] $(msg "Stopped" "중지됨")"
		exit 0
	fi

	case "$code" in
		0)
			echo "[Skopos] $(msg "Exited normally" "정상 종료")"
			exit 0
			;;
		"$RESTART_CODE")
			echo "[Skopos] $(msg "Restart requested (applying update)…" "다시 시작 요청 (업데이트 적용 중)…")"
			delay=1
			;;
		*)
			echo "[Skopos] $(msg "Exited with code $code; restarting in ${delay}s" "종료 코드 $code; ${delay}초 후 다시 시작")" >&2
			sleep "$delay"
			delay=$((delay * 2))
			[ "$delay" -gt "$MAX_DELAY" ] && delay=$MAX_DELAY
			;;
	esac
done
