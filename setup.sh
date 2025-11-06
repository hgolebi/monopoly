#!/bin/bash

# Skrypt do instalacji zależności, pobrania i uruchomienia projektu Go na nowej maszynie Linux (Debian/Ubuntu).

# Zatrzymaj skrypt, jeśli jakakolwiek komenda zakończy się błędem.
set -e

# --- Konfiguracja ---
REPO_URL="https://github.com/hgolebi/monopoly.git"
PROJECT_DIR="monopoly"
MAIN_FILE="main.go"
GO_VERSION="1.23.0"

# --- Kroki ---

# 1. Instalacja najnowszej wersji Go
echo ">>> Krok 1: Instalacja Go w wersji $GO_VERSION..."
if ! command -v go &> /dev/null || ! go version | grep -q "$GO_VERSION"; then
    echo "Pobieranie Go $GO_VERSION..."
    wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go.tar.gz
    rm go.tar.gz

    # --- POCZĄTEK ZMIAN ---
    # Sprawdź, czy ścieżka Go jest już w .bashrc, aby uniknąć duplikatów
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo "Dodawanie Go do PATH na stałe w ~/.bashrc..."
        echo '' >> ~/.bashrc
        echo '# Set Go path' >> ~/.bashrc
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi

    # Zastosuj zmiany w bieżącej sesji
    export PATH=$PATH:/usr/local/go/bin
    # --- KONIEC ZMIAN ---

    echo ">>> Go $GO_VERSION zostało zainstalowane."
else
    echo "Go w wersji $GO_VERSION jest już zainstalowane."
fi
go version
echo ""

# 2. Instalacja Git
echo ">>> Krok 2: Instalacja Git..."
sudo apt-get update
sudo apt-get install -y git
echo ">>> Git został zainstalowany."
echo ""

# 3. Klonowanie repozytorium projektu
echo ">>> Krok 3: Klonowanie repozytorium z $REPO_URL..."
if [ -d "$PROJECT_DIR" ]; then
    echo "Folder projektu '$PROJECT_DIR' już istnieje. Usuwam i klonuję ponownie."
    rm -rf "$PROJECT_DIR"
fi
git clone "$REPO_URL" "$PROJECT_DIR"
echo ">>> Repozytorium zostało sklonowane."
echo ""

# 4. Przejście do folderu projektu
cd "$PROJECT_DIR"
echo ">>> Przechodzę do folderu: $(pwd)"
echo ""

# 5. Pobranie zależności Go
echo ">>> Krok 4: Pobieranie modułów Go z pliku go.mod..."
go mod tidy
go mod download
echo ">>> Moduły zostały pobrane."
echo ""

# 6. Uruchomienie programu
echo ">>> Krok 5: Uruchamianie programu ($MAIN_FILE)..."
echo "--------------------------------------------------"
go run "$MAIN_FILE"
echo "--------------------------------------------------"
echo ">>> Program zakończył działanie."

exit 0