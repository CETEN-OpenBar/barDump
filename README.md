# BarDump 

BarDump est une application web de type "Wrapped" (à l'instar de Spotify Wrapped) conçue spécifiquement pour analyser et présenter les statistiques de consommation des clients d'un bar sous forme de stories (diaporama interactif).

Le projet s'appuie sur les données d'encaissement et de transactions d'une base MongoDB, génère des statistiques par utilisateur (total dépensé, produits favoris, volume d'achats mensuels, classement, etc.), et les expose sur une interface web fluide et animée.

## 🏗 Architecture du Projet

Ce dépôt est un monorepo comprenant deux parties principales :

1. **Backend / Data Processing (Python)** :
   Situé dans le dossier `backend/scripts`, ce module est responsable de la récupération et du traitement des données.
   - `export_db.py` : Se connecte à la base MongoDB pour exporter les collections requises.
   - `process_stats.py` : Ingère les transactions exportées, génère les classements et agrège les données de chaque utilisateur pour le rendu final.

2. **Frontend & API (SvelteKit)** :
   Situé dans le dossier `wrapped-app`.
   - **Vue Publique (`/dump/[year]/[user_id]`)** : L'interface utilisateur finale. Un carrousel plein écran affichant les statistiques de l'utilisateur sous forme de diapositives avec des transitions fluides.
   - **Vue Admin (`/admin`)** : Interface de gestion permettant de créer de nouvelles campagnes ("dumps"), de définir les logos, et de déclencher manuellement les scripts de mise à jour des données (en faisant le pont avec les scripts Python).

## 🚀 Prérequis

- **Node.js** (v20+) & `npm` ou `pnpm`
- **Python** (3.8+)
- **MongoDB** (Accès à la base de données source du bar)
- **Docker** (Optionnel, pour le déploiement)

## 🛠 Variables d'Environnement

Pour des raisons de sécurité, les identifiants de la base de données ne sont pas inclus dans le code. Vous devez définir la variable d'environnement suivante avant d'exécuter l'extraction des données :

```bash
MONGODB_URI="mongodb://utilisateur:motdepasse@hote:port/?authSource=database"
MONGODB_DB_NAME="nom_de_la_base" # "bar" par défaut
ORIGIN=http://localhost:3000
```

## 💻 Développement Local

### 1. Préparation du Backend (Python)

```bash
cd backend/scripts
# Création d'un environnement virtuel (recommandé)
python3 -m venv venv
source venv/bin/activate
# Installation des dépendances
pip install -r requirements.txt
```

### 2. Démarrage de l'Application (SvelteKit)

```bash
cd wrapped-app
# Installation des dépendances du frontend
npm install
# Lancement du serveur de développement
npm run dev
```

L'application sera accessible sur `http://localhost:5173`. Vous pouvez ensuite accéder à la page d'administration via `http://localhost:5173/admin` pour uploader vos logos et générer un dump.

## 🐳 Déploiement avec Docker

Le projet inclut un `Dockerfile` multi-stage, optimisé pour la production. Il compile l'application SvelteKit puis crée une image Node.js enrichie d'un environnement Python pour exécuter les scripts de traitement de données.

1. **Construire l'image :**
```bash
docker build -t bardump .
```

2. **Lancer le conteneur :**
```bash
docker run -d \
  -p 3000:3000 \
  -e MONGODB_URI="votre_uri_mongo" \
  -v $(pwd)/data:/app/data \
  --name bardump-app \
  bardump
```

Les données (fichiers JSON bruts et traités) seront stockées dans le dossier `/app/data` (mappé au dossier local `./data` pour la persistance).

## 🗂 Structure du Dépôt

```
.
├── Dockerfile                  # Configuration pour le déploiement
├── README.md                   # Ce fichier
├── backend/
│   └── scripts/
│       ├── requirements.txt    # Dépendances Python (pymongo, etc.)
│       ├── export_db.py        # Extraction depuis MongoDB
│       └── process_stats.py    # Logique de calcul des statistiques
└── wrapped-app/
    ├── package.json
    ├── svelte.config.js
    ├── tailwind.config.ts      # Configuration du framework CSS
    ├── static/                 # Fichiers statiques et logos
    └── src/
        ├── lib/
        │   └── components/     # Composants Svelte (les Slides du Wrapped)
        └── routes/
            ├── admin/          # Dashboard d'administration
            ├── api/            # API backend Node.js
            └── dump/           # Vue principale (le Wrapped)
```
