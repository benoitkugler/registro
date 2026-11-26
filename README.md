# Registro, un système d'inscription et de gestion associative

## Configuration

La configuration pour une association se fait par 3 moyens :

- le fichier server/config/asso.go
- les variables d'environnement :

  | Variable                                | Description                                                                                                                             | Version             |
  | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------------------- |
  | `PAYREXX_INSTANCE`                      | instance Payrexx, sans `https://` (quand le fournisseur `payrexx` est sélectionné)                                                      | v0.14.0             |
  | `PAYREXX_API_KEY`                       | clé API Payrexx (quand le fournisseur `payrexx` est sélectionné)                                                                        | v0.14.0             |
  | `STRIPE_KEY`                            | clé API Stripe (quand le fournisseur `stripe` est sélectionné)                                                                          | v0.1.1 ou antérieur |
  | `STRIPE_WEBHOOK`                        | clé secrète de vérification des notifications Stripe (quand le fournisseur `stripe` est sélectionné)                                    | v0.1.1 ou antérieur |
  | `ANCV_SHOP_ID`                          | identifiant technique de la boutique ANCV (`merchant.shopId`) ; intégration directe, sans `serviceProviderId`                            | v0.15.0             |
  | `ANCV_SEAL_KEY`                         | clé de scellement HMAC-SHA256 du commerçant (PTL)                                                                                         | v0.15.0             |
  | `ANCV_SEAL_VERSION`                     | version de la clé de scellement, placée dans l'en-tête `ANCV-Security`                                                                   | v0.15.0             |
  | `ANCV_PROD`                             | `TRUE` pour utiliser la production ANCV, sinon (ou si absent) l'environnement de recette                                                 | v0.15.0             |
  | `ASSO_MAIL_NOTIFICATION_FROM_DIRECTEUR` | adresse de notification pour les messages automatiques envoyés quand un directeur demande l'attention sur un dossier                    | v0.13.0             |
  | `COMPTA_ADMIN_PASSWORD`                 | mot de passe administrateur de l'espace compta : il donne accès au suivi de tous les séjours                                            | v0.13.0             |
  | `COMPTA_ADMIN_MAIL`                     | adresse mail de l'administrateur de l'espace compta, sur lequel est envoyé une notification quand le directeur a terminé sa déclaration | v0.13.0             |
  | `COMPTA_GOOGLE_DRIVE_FOLDER_ID`         | identifiant du dossier Google Drive partagé avec le compte de service (optionnel par défaut, requis avec le flag -checkComptaDrive)     | v0.13.0             |
  | `COMPTA_GOOGLE_DRIVE_KEY_FILE`          | chemin vers le fichier de clé JSON du compte de service Google Drive (optionnel par défaut, requis avec le flag -checkComptaDrive)      | v0.13.0             |
  | `COMPTA_FILES_DIR`                      | répertoire local des fichiers des justificatifs déposés sur l'espace comptabilité des séjours                                           | v0.13.0             |
  | `HELLOASSO_HORIZON_KEY`                 | clé d'authentification des requêtes Horizon entrantes                                                                                   | v0.12.3             |
  | `HELLOASSO_RELAIS_KEY`                  | clé d'authentification des réponses envoyées à Horizon                                                                                  | v0.12.3             |
  | `ASSO_MAIL_FONDS_SOUTIEN`               | adresse de notification pour le fonds de soutien                                                                                        | v0.9.2              |
  | `ASSO_MAIL_SAUVEGARDE`                  | adresse recevant une copie des courriels envoyés                                                                                        | v0.9.0              |
  | `DONS_PASSWORD`                         | mot de passe d'accès à l'espace dons                                                                                                    | v0.9.0              |
  | `HELLOASSO_SLUG`                        | identifiant (slug) de l'association sur HelloAsso                                                                                       | v0.9.0              |
  | `HELLOASSO_SANDBOX`                     | `TRUE` pour utiliser l'environnement de test HelloAsso                                                                                  | v0.9.0              |
  | `MODELE_RECU_FISCAL`                    | chemin vers le modèle PDF de reçu fiscal (vide pour désactiver l'édition des reçus)                                                     | v0.9.0              |
  | `IMMICH_BASE_URL`                       | URL de base de l'instance Immich (schéma et hôte, sans chemin)                                                                          | v0.8.0              |
  | `IMMICH_API_KEY`                        | clé API Immich                                                                                                                          | v0.8.0              |
  | `MEDIA_DIR`                             | répertoire externe optionnel exposé comme fichiers statiques                                                                            | v0.7.3              |
  | `FOND_SOUTIEN_PASSWORD`                 | mot de passe d'accès à l'espace « fonds de soutien »                                                                                    | v0.3.0              |
  | `FILES_DIR`                             | répertoire de stockage des fichiers utilisateurs (obligatoire)                                                                          | v0.2.0              |
  | `ASSETS_DIR`                            | répertoire des ressources statiques en lecture seule, utilisées dans les modèles HTML (obligatoire)                                     | v0.2.0              |
  | `CACHE_DIR`                             | répertoire du cache des polices (obligatoire)                                                                                           | v0.2.0              |
  | `ASSO`                                  | association à configurer (`acve` ou `repere`)                                                                                           | v0.1.1 ou antérieur |
  | `ASSO_BANK_IBAN`                        | IBAN des comptes bancaires, séparés par des virgules (un par compte de l'association)                                                   | v0.1.1 ou antérieur |
  | `SERVER_KEY`                            | clé de chiffrement des données                                                                                                          | v0.1.1 ou antérieur |
  | `BACKOFFICE_PASSWORD`                   | mot de passe d'accès au backoffice                                                                                                      | v0.1.1 ou antérieur |
  | `DIRECTEURS_PASSWORD`                   | mot de passe global d'accès à l'espace directeurs (tous les camps)                                                                      | v0.1.1 ou antérieur |
  | `SMTP_HOST`                             | hôte du serveur SMTP                                                                                                                    | v0.1.1 ou antérieur |
  | `SMTP_USER`                             | utilisateur SMTP (adresse mail valide)                                                                                                  | v0.1.1 ou antérieur |
  | `SMTP_PASSWORD`                         | mot de passe SMTP                                                                                                                       | v0.1.1 ou antérieur |
  | `SMTP_PORT`                             | port du serveur SMTP                                                                                                                    | v0.1.1 ou antérieur |
  | `DB_HOST`                               | hôte de la base de données PostgreSQL                                                                                                   | v0.1.1 ou antérieur |
  | `DB_USER`                               | utilisateur PostgreSQL                                                                                                                  | v0.1.1 ou antérieur |
  | `DB_PASSWORD`                           | mot de passe PostgreSQL                                                                                                                 | v0.1.1 ou antérieur |
  | `DB_NAME`                               | nom de la base de données                                                                                                               | v0.1.1 ou antérieur |
  | `DB_PORT`                               | port PostgreSQL (facultatif, 5432 par défaut)                                                                                           | v0.1.1 ou antérieur |
  | `HELLOASSO_ID`                          | identifiant de l'API HelloAsso                                                                                                          | v0.1.1 ou antérieur |
  | `HELLOASSO_SECRET`                      | secret de l'API HelloAsso                                                                                                               | v0.1.1 ou antérieur |
  | `IP`                                    | adresse IP d'écoute du serveur, en production                                                                                           | v0.1.1 ou antérieur |
  | `PORT`                                  | port d'écoute du serveur, en production                                                                                                 | v0.1.1 ou antérieur |

- l'object `colorScheme` du fichier registro-web/src/vuetify.ts

## Paiement par Chèque-Vacances Connect (ANCV)

L'intégration est directe (commerçant/PTL, sans intermédiaire de service).
Les variables `ANCV_*` ne sont lues que lorsque l'association active
`SupportCVCo` (actuellement ACVE). Le montant est exprimé en centimes d'euros,
seule devise acceptée par l'ANCV.

Points d'entrée :

- `POST /api/v1/espaceperso/paiement/ancv` : initialise la transaction et
  soumet le paiement au bénéficiaire ;
- `POST /api/v1/ancv/notification` : webhook public (sans JWT ni filtrage d'IP).
  Son état est revérifié auprès de l'ANCV avant tout enregistrement.
  L'URL est construite depuis l'hôte de la requête, comme pour Stripe.

### Références ANCV

Les commentaires de `server/onlinepay/ancv/` et de
`server/controllers/espaceperso/ancv.go` citent le document
`Kit_Integration_par_API_du_CVCo.pdf` (v1.08, © Worldline) par section et page.
Les sections utilisées sont principalement :

| Section | Pages | Sujet |
| ------- | ----- | ----- |
| §4.2 | 13 | environnements de recette et de production |
| §4.3.1, §5.3.2 | 13, 52-53 | en-têtes et calcul du sceau HMAC-SHA256 |
| §4.5.1, §4.5.2, §4.5.3 | 17-25 | initialisation, mise en paiement, demande d'état |
| §5.1 | 45-46 | webhooks, rejeux et contraintes de sécurité |
| §6.2.2 à §6.2.5 | 56-69 | dictionnaire des données échangées |
| §6.3 | 69-70 | correspondance des codes d'erreur |
| §6.7 | 81 | vecteur de test du calcul de sceau |

Le kit ne fournit pas d'API pour simuler des notifications (`§6.9 Cahier de
recette - A venir`). En recette, les notifications sont générées par de vraies
transactions validées depuis l'application bénéficiaire (§6.10, document DR3).

