-- Code genererated by gomacro/generator/sql. DO NOT EDIT.
CREATE TABLE dons (
    Id serial PRIMARY KEY,
    Valeur Montant NOT NULL,
    ModePaiement smallint CHECK (ModePaiement IN (0, 1, 2, 3, 4, 5)) NOT NULL,
    Date date NOT NULL,
    Affectation text NOT NULL,
    Details text NOT NULL,
    Remercie boolean NOT NULL,
    IdPaiementHelloasso text NOT NULL
);

-- constraints
