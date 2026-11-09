# Registro, un système d'inscription et de gestion associative

## Configuration

La configuration pour une association se fait par 3 moyens :

- le fichier server/config/asso.go
- les variables d'environnement :
  - ASSO
  - ASSO_BANK_IBAN
  - ASSO_BANK_IBAN
  - ASSO_MAIL_SAUVEGARDE
  - ASSO_MAIL_FONDS_SOUTIEN
  - SERVER_KEY
  - BACKOFFICE_PASSWORD
  - FOND_SOUTIEN_PASSWORD
  - DIRECTEURS_PASSWORD
  - DONS_PASSWORD
  - FILES_DIR
  - ASSETS_DIR
  - CACHE_DIR
  - MEDIA_DIR
  - SMTP_HOST
  - SMTP_USER
  - SMTP_PASSWORD
  - SMTP_PORT
  - DB_HOST
  - DB_USER
  - DB_PASSWORD
  - DB_NAME
  - DB_PORT
  - STRIPE_KEY
  - STRIPE_WEBHOOK
  - HELLOASSO_SLUG
  - HELLOASSO_ID
  - HELLOASSO_SECRET
  - HELLOASSO_SANDBOX
  - HELLOASSO_HORIZON_KEY
  - HELLOASSO_RELAIS_KEY
  - MODELE_RECU_FISCAL
  - IMMICH_BASE_URL
  - IMMICH_API_KEY
  - COMPTA_ADMIN_PASSWORD
  - COMPTA_ADMIN_MAIL
  - COMPTA_GOOGLE_DRIVE_FOLDER_ID
  - COMPTA_GOOGLE_DRIVE_KEY_FILE
  - COMPTA_FILES_DIR
- l'object `colorScheme` du fichier registro-web/src/vuetify.ts
