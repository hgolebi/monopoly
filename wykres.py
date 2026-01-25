import pandas as pd
import matplotlib.pyplot as plt
import re

def wczytaj_i_przygotuj_dane(nazwa_pliku):
    """
    Wczytuje dane, filtruje tylko dla 'bracket133' i parsuje je do DataFrame.
    (Ta funkcja jest identyczna jak w poprzedniej odpowiedzi, aby zachować kompletność)
    """
    try:
        with open(nazwa_pliku, 'r') as file:
            linie = file.readlines()
    except FileNotFoundError:
        print(f"Błąd: Nie znaleziono pliku {nazwa_pliku}")
        return None

    dane = []
    rgx_games_played = re.compile(r"Games played:\s+(\d+)")
    rgx_bracket = re.compile(r"bracket133:\s+AvgScore=(\d+)\s+Wins=\d+\s+Draws=\d+")

    current_games_played = 0

    for linia in linie:
        m_gp = rgx_games_played.search(linia)
        if m_gp:
            current_games_played = int(m_gp.group(1))
            continue

        m_bracket = rgx_bracket.search(linia)
        if m_bracket and current_games_played > 0:
            avg_score = int(m_bracket.group(1))

            dane.append({
                'Games_Played': current_games_played,
                'AvgScore': avg_score
            })

    df = pd.DataFrame(dane)
    return df

def stworz_wykres_pudelkowy(df):
    """
    Tworzy wykres pudełkowy dla AvgScore pogrupowany według Games_Played.
    """
    if df is None:
        return

    # Upewnienie się, że 'Games_Played' jest traktowane jako kategoria (zmieniamy kolejność)
    df['Games_Played'] = df['Games_Played'].astype('category')

    # Sortowanie kategorii, aby były w odpowiedniej kolejności (100, 1000, 2000)
    df['Games_Played'] = df['Games_Played'].cat.set_categories([100, 1000, 2000, 10000])

    plt.figure(figsize=(9, 7))

    md_props = dict(
        color='red',       # Kolor linii mediany
        linewidth=2.0        # Grubość linii mediany (np. 3.0)
    )

    # Tworzenie wykresu pudełkowego z Pandas/Matplotlib
    # `column` to wartość, którą badamy (`AvgScore`)
    # `by` to kolumna, wg której grupujemy (`Games_Played`)
    df.boxplot(column='AvgScore', by='Games_Played', ax=plt.gca(), patch_artist=True, medianprops=md_props)

    # Dostosowanie tytułów i etykiet
    plt.suptitle('') # Usuwamy automatyczny tytuł `boxplot(by=...)`
    plt.title('', fontsize=15)
    plt.xlabel('Liczba rozegranych gier', fontsize=12)
    plt.ylabel('Średnia liczba punktów', fontsize=12)
    plt.grid(axis='y', linestyle='--', alpha=0.7)

    plt.tight_layout()
    plt.show()

# --- Główna część programu ---
if __name__ == "__main__":
    NAZWA_PLIKU = 'wyniki_testow/games_per_epoch'  # Plik z danymi

    # 1. Wczytanie i przetworzenie danych
    dane_df = wczytaj_i_przygotuj_dane(NAZWA_PLIKU)

    if dane_df is not None:
        # 2. Tworzenie i wyświetlanie wykresu pudełkowego
        stworz_wykres_pudelkowy(dane_df)