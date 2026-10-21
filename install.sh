#!/usr/bin/env bash
# Skopos install script: ① all-in-Docker  ② build and run locally
#
# Messages are English by default. For Korean, set SKOPOS_LANGUAGE=ko
# (SKOPOS_LANGUAGE=ko also works) before running, e.g. SKOPOS_LANGUAGE=ko ./install.sh
set -euo pipefail
umask 077
cd "$(cd "$(dirname "$0")" && pwd)"

LANG_SEL="${SKOPOS_LANGUAGE:-${SKOPOS_LANGUAGE:-en}}"
case "$LANG_SEL" in ko|ko_*|ko-*|KO) LANG_SEL=ko ;; *) LANG_SEL=en ;; esac
# msg EN KO → print the string for the selected language.
msg(){ if [ "$LANG_SEL" = ko ]; then printf '%s' "$2"; else printf '%s' "$1"; fi; }

info(){ printf '\033[36m[*]\033[0m %s\n' "$*"; }
ok(){   printf '\033[32m[+]\033[0m %s\n' "$*"; }
warn(){ printf '\033[33m[!]\033[0m %s\n' "$*"; }
die(){  printf '\033[31m[x]\033[0m %s\n' "$*" >&2; exit 1; }
ask(){  local p="$1" d="${2:-}" a; read -rp "$p${d:+ [$d]}: " a; echo "${a:-$d}"; }
rand(){ head -c 18 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 24; }

# Protect before writing. Windows chmod does not remove inherited ACL grants.
# Destination directories must belong to the operator; this is not a sandbox
# against concurrent replacement of path components by another process.
private_credential(){
  [ ! -L "$1" ] && { [ ! -e "$1" ] || [ -f "$1" ]; } || die "$(msg 'Credential destination must be a regular file' '자격 증명 대상은 일반 파일이어야 합니다')"
  case "$(uname -s)" in
    MINGW*|MSYS*|CYGWIN*)
      command -v powershell.exe >/dev/null 2>&1 || die "$(msg 'Windows PowerShell is required to protect credential files' '자격 증명 파일을 보호하려면 Windows PowerShell이 필요합니다')"
      powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden -File "$(cygpath -w "$PWD/private-credential.ps1")" -Path "$(cygpath -w "$PWD/$1")" || die "$(msg 'Could not protect credential file' '자격 증명 파일을 보호할 수 없습니다')"
      ;;
    *)
      if [ ! -e "$1" ]; then (set -C; : > "$1"); fi
      chmod 600 "$1"
      ;;
  esac
}

# ── docker environment detection / auto-install ─────────────
ensure_docker(){
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    ok "$(msg 'Found docker and docker compose' 'docker와 docker compose를 감지했습니다')"; return
  fi
  warn "$(msg 'docker / docker compose not found' 'docker / docker compose를 찾지 못했습니다')"
  case "$(uname -s)" in
    Linux)
      if [ "$(ask "$(msg 'Install Docker automatically? (y/n)' 'Docker를 자동으로 설치할까요? (y/n)')" y)" = y ]; then
        curl -fsSL https://get.docker.com | sh
        sudo usermod -aG docker "$USER" || true
        ok "$(msg 'Docker installed (group change needs re-login to use without sudo)' 'Docker 설치 완료 (그룹 변경은 재로그인 후 sudo 없이 적용)')"
      else
        die "$(msg 'Please install docker yourself and retry' 'docker를 직접 설치한 뒤 다시 시도하세요')"
      fi ;;
    Darwin) die "$(msg 'On macOS install Docker Desktop: https://www.docker.com/products/docker-desktop/' 'macOS에서는 Docker Desktop을 설치하세요: https://www.docker.com/products/docker-desktop/')" ;;
    *)      die "$(msg 'Please install docker yourself and retry' 'docker를 직접 설치한 뒤 다시 시도하세요')" ;;
  esac
}

# ── ① all-in-Docker ─────────────────────────────────────────
install_docker(){
  [ -s tls/server.pem ] && [ -s tls/server-key.pem ] || die "$(msg 'Docker management requires HTTPS: place a trusted certificate in tls/server.pem and its matching key in tls/server-key.pem (see README).' 'Docker 관리는 HTTPS가 필요합니다: 신뢰할 수 있는 인증서를 tls/server.pem에, 일치하는 키를 tls/server-key.pem에 저장하세요 (README 참고).')"
  ensure_docker
  [ ! -L .env ] || die "$(msg 'Refusing a symlinked .env credential file' '심볼릭 링크 .env 자격 증명 파일을 거부합니다')"
  if [ ! -f .env ]; then
    [ -f .env.example ] || die "$(msg 'Missing .env.example template' '.env.example 템플릿이 없습니다')"
    private_credential .env
    local pw key
    pw="$(ask "$(msg 'Postgres password (Enter for a random one)' 'Postgres 비밀번호 (Enter 시 무작위 생성)')" "$(rand)")"
    key="$(ask "$(msg 'ANTHROPIC_API_KEY (optional, can set later in the UI)' 'ANTHROPIC_API_KEY (선택, 나중에 UI에서 설정 가능)')" '')"
    # Write through the protected file, without replacement/backup files or
    # treating credential characters as sed syntax or command-line arguments.
    while IFS= read -r line || [ -n "$line" ]; do
      case "$line" in
        POSTGRES_PASSWORD=*) printf 'POSTGRES_PASSWORD=%s\n' "$pw" ;;
        ANTHROPIC_API_KEY=*) printf 'ANTHROPIC_API_KEY=%s\n' "$key" ;;
        *) printf '%s\n' "$line" ;;
      esac
    done < .env.example > .env
    ok "$(msg '.env generated (POSTGRES_PASSWORD set)' '.env 생성 완료 (POSTGRES_PASSWORD 설정됨)')"
  else
    private_credential .env
    info "$(msg 'Reusing existing .env' '기존 .env를 그대로 사용합니다')"
  fi
  info "$(msg 'Building Skopos from this checkout…' '현재 소스로 Skopos를 빌드합니다…')"
  docker compose build skopos
  docker compose up -d
  ok "$(msg 'Started → https://localhost:8787' '시작 완료 → https://localhost:8787')"
  info "$(msg 'View logs: docker compose logs -f skopos' '로그 보기: docker compose logs -f skopos')"
}

# ── ② build and run locally ─────────────────────────────────
install_local(){
  echo "$(msg 'Database setup:' '데이터베이스 설치 방식:')"
  echo "  1) $(msg 'Connect to an existing PostgreSQL' '기존 PostgreSQL에 연결')"
  echo "  2) $(msg 'Start a PostgreSQL with Docker (needs docker)' 'Docker로 PostgreSQL 하나 띄우기 (docker 필요)')"
  case "$(ask "$(msg 'Choose' '선택')" 1)" in
    2)
      ensure_docker
      local pw; pw="$(ask "$(msg 'Postgres password (Enter for a random one)' 'Postgres 비밀번호 (Enter 시 무작위 생성)')" "$(rand)")"
      docker run -d --name skopos-pg -p 5432:5432 \
        -e POSTGRES_USER=artex -e POSTGRES_PASSWORD="$pw" -e POSTGRES_DB=artex \
        -v skopos-pg:/var/lib/postgresql/data postgres:16-alpine
      DB_HOST=127.0.0.1 DB_PORT=5432 DB_USER=artex DB_PASS="$pw" DB_NAME=artex DB_SSL=disable ;;
    *)
      DB_HOST="$(ask "$(msg 'Database host' '데이터베이스 주소')" 127.0.0.1)"
      DB_PORT="$(ask "$(msg 'Port' '포트')" 5432)"
      DB_USER="$(ask "$(msg 'User' '계정')" artex)"
      DB_PASS="$(ask "$(msg 'Password' '비밀번호')" '')"
      DB_NAME="$(ask "$(msg 'Database name' '데이터베이스 이름')" artex)"
      DB_SSL="$(ask "$(msg 'sslmode (disable/require)' 'sslmode (disable/require)')" disable)" ;;
  esac

  # Generate config.json
  [ ! -L config.json ] || die "$(msg 'Refusing a symlinked config.json credential file' '심볼릭 링크 config.json 자격 증명 파일을 거부합니다')"
  private_credential config.json
  cat > config.json <<JSON
{
  "database": {
    "host": "${DB_HOST}",
    "port": ${DB_PORT},
    "user": "${DB_USER}",
    "password": "${DB_PASS}",
    "dbname": "${DB_NAME}",
    "sslmode": "${DB_SSL}"
  }
}
JSON
  ok "$(msg 'config.json generated' 'config.json 생성 완료')"

  # go environment check
  command -v go >/dev/null 2>&1 || die "$(msg 'Go not found, install Go (>=1.26): https://go.dev/dl/' 'Go를 찾지 못했습니다. Go(>=1.26)를 먼저 설치하세요: https://go.dev/dl/')"
  ok "Go: $(go version)"

  # Embedding the frontend needs node to produce the static output
  if command -v npm >/dev/null 2>&1; then
    info "$(msg 'Building the frontend static output…' '프런트엔드 정적 산출물을 빌드합니다…')"
    ( cd web && npm ci && npm run build:static )
    rm -rf server/webui/dist && cp -r web/out server/webui/dist
    info "$(msg 'Compiling the single binary with the frontend embedded…' '프런트엔드가 내장된 단일 바이너리를 컴파일합니다…')"
    CGO_ENABLED=0 go build -tags embedui -trimpath -o skopos ./cmd/skopos
  else
    warn "$(msg 'npm not found: building the backend WITHOUT the embedded frontend (run npm run dev for the frontend)' 'npm을 찾지 못함: 프런트엔드 미내장 백엔드를 빌드합니다 (프런트엔드는 따로 npm run dev)')"
    CGO_ENABLED=0 go build -o skopos ./cmd/skopos
  fi
  ok "$(msg 'Build complete → ./skopos' '컴파일 완료 → ./skopos')"

  info "$(msg 'Starting… (Ctrl-C to quit)' '시작합니다… (Ctrl-C로 종료)')"
  ./start.sh
}

echo "=============================="
echo "  $(msg 'Skopos install' 'Skopos 설치')"
echo "  1) $(msg 'All-in-Docker install' '전부 Docker로 설치')"
echo "  2) $(msg 'Local run (go build)' '로컬 실행 (go 컴파일)')"
echo "=============================="
case "$(ask "$(msg 'Choose' '선택')" 1)" in
  1) install_docker ;;
  2) install_local ;;
  *) die "$(msg 'Invalid choice' '잘못된 선택')" ;;
esac
