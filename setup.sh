#!/bin/bash

# Skrypt do instalacji zależności, pobrania i uruchomienia projektu Go na nowej maszynie Linux (Debian/Ubuntu).

# Zatrzymaj skrypt, jeśli jakakolwiek komenda zakończy się błędem.
set -e

# --- Konfiguracja ---
# !!! WAŻNE: Zmień ten adres URL na adres swojego repozytorium Git !!!
REPO_URL="https://github.com/hgolebi/monopoly.git"
# Nazwa folderu, do którego zostanie sklonowany projekt.
PROJECT_DIR="monopoly"
# Główny plik do uruchomienia.
MAIN_FILE="main.go"

# --- Kroki ---

# 1. Instalacja podstawowych narzędzi (Go i Git)
echo ">>> Krok 1: Instalacja Go i Git..."
sudo apt-get update
sudo apt-get install -y golang-go git
echo ">>> Go i Git zostały zainstalowane."
echo ""

# Weryfikacja instalacji Go
go version
echo ""

# 2. Klonowanie repozytorium projektu
echo ">>> Krok 2: Klonowanie repozytorium z $REPO_URL..."
if [ -d "$PROJECT_DIR" ]; then
    echo "Folder projektu '$PROJECT_DIR' już istnieje. Pomijam klonowanie."
else
    git clone "$REPO_URL" "$PROJECT_DIR"
fi
echo ">>> Repozytorium zostało sklonowane."
echo ""

# 3. Przejście do folderu projektu
cd "$PROJECT_DIR"
echo ">>> Przechodzę do folderu: $(pwd)"
echo ""

# 4. Pobranie zależności Go
echo ">>> Krok 3: Pobieranie modułów Go z pliku go.mod..."
go mod tidy
go mod download
echo ">>> Moduły zostały pobrane."
echo ""

# 5. Uruchomienie programu
echo ">>> Krok 4: Uruchamianie programu ($MAIN_FILE)..."
echo "--------------------------------------------------"
go run "$MAIN_FILE"
echo "--------------------------------------------------"
echo ">>> Program zakończył działanie."

exit 0