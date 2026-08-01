# BarDump 

BarDump est une application web de type "Wrapped" (à l'instar de Spotify Wrapped) conçue spécifiquement pour analyser et présenter les statistiques de consommation des clients d'un bar sous forme de stories (diaporama interactif).

Le projet s'appuie sur les données d'encaissement et de transactions d'une base MongoDB, génère des statistiques par utilisateur (total dépensé, produits favoris, volume d'achats mensuels, classement, etc.), et les expose sur une interface web fluide et animée.

##  Architecture du Projet

Ce dépôt est un monorepo comprenant deux parties principales :

1. **Backend / Data Processing (Go API)** :
   Situé dans le dossier `backend/go-api`, ce module est responsable de la récupération et du traitement massif des données, ainsi que de l'envoi des e-mails.
   - `internal/handlers/` : Parse les requêtes HTTP envoyées par le backend SvelteKit.
   - `internal/services/exporter.go` : Se connecte à la base MongoDB pour exporter les transactions et les comptes.
   - `internal/services/processor.go` : Ingère les transactions exportées, génère les classements et agrège les données de chaque utilisateur pour le rendu final.
   - `internal/services/mailer.go` : Gère l'envoi massif de mails via SMTP pour prévenir les utilisateurs avec un suivi de progression en SSE.

2. **Frontend & Admin API (SvelteKit)** :
   Situé dans le dossier `wrapped-app`.
   - **Vue Publique (`/dump/[year]/[user_id]`)** : L'interface utilisateur finale. Un carrousel plein écran affichant les statistiques de l'utilisateur sous forme de diapositives avec des transitions fluides.
   - **Vue Admin (`/admin`)** : Interface de gestion permettant de créer de nouvelles campagnes ("dumps"), uploader des logos, et streamer (SSE) l'avancement des générations en faisant le pont avec l'API Go.

##  Prérequis

- **Node.js** (v20+) & `npm`
- **Go** (1.21+)
- **MongoDB** (Accès à la base de données source du bar)
- **Docker** (Optionnel, fortement recommandé pour le déploiement)

## 🛠 Variables d'Environnement

Le projet utilise un fichier `.env` situé dans `wrapped-app/.env` (qui est aussi lu par l'API Go en développement local). 

```bash
MONGODB_URI="mongodb://utilisateur:motdepasse@hote:port/?authSource=database"
MONGODB_DB_NAME="nom_de_la_base" # "bar" par défaut
ORIGIN=http://localhost:3000
SMTP_EMAIL="votre_adresse_mail@gmail.com"
SMTP_PASSWORD="votre_mot_de_passe_d_application"
BASE_URL="https://votre-domaine.fr"
```

##  Développement Local

### 1. Démarrage de l'API Go

```bash
cd backend/go-api
go mod download
go run ./cmd/server
```
L'API Go démarrera sur `http://localhost:8080`.

### 2. Démarrage de l'Application SvelteKit

Dans un nouveau terminal :
```bash
cd wrapped-app
npm install
npm run dev
```

L'application sera accessible sur `http://localhost:5173`. L'administration se trouve sur `/admin`.

##  Déploiement avec Docker

Le projet inclut un `Dockerfile` multi-stage. Il compile l'API Go puis l'application SvelteKit dans une image légère basée sur Node.

1. **Construire l'image :**
```bash
docker build -t bardump .
```

2. **Lancer le conteneur :**
```bash
docker run -d \
  -p 3000:3000 \
  -e MONGODB_URI="votre_uri_mongo" \
  -e SMTP_EMAIL="votre_mail" \
  -e SMTP_PASSWORD="votre_password" \
  -v $(pwd)/data:/app/data \
  --name bardump-app \
  bardump
```

L'application expose automatiquement le port `3000`. Les données générées (les statistiques json et les configurations) seront stockées dans le dossier `/app/data` (mappé au dossier local `./data` pour la persistance).

##  Structure du Dépôt

```text
.
├── Dockerfile                  # Configuration pour le build multi-stage Docker (Go + Node)
├── docker-compose.yml          # Modèle optionnel pour le déploiement
├── README.md                   # Ce fichier
├── backend/
│   └── go-api/
│       ├── cmd/server/main.go  # Entrée du serveur Echo (Port 8080)
│       └── internal/           # Logique d'export, de processing et d'emails (SSE)
└── wrapped-app/
    ├── package.json
    ├── svelte.config.js
    ├── tailwind.config.ts      
    ├── static/                 # Fichiers statiques et logos
    └── src/
        ├── lib/
        │   └── components/     # Composants Svelte (les Slides du Wrapped)
        └── routes/
            ├── admin/          # Dashboard d'administration (Appels SSE vers /api/dumps)
            ├── api/            # Routeurs proxy SvelteKit -> API Go
            └── dump/           # Vue principale (le Wrapped)
```
