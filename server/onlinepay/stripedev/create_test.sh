# Forward the events to out custom server listening on stripe_test
# stripe listen --all-snapshot --forward-to localhost:1323/stripe_test -l
stripe trigger checkout.session.completed --override price:currency=eur  --override checkout_session:currency=eur --override checkout_session:metadata.id_dossier=1 &&
stripe trigger checkout.session.completed --override price:currency=chf  --override checkout_session:currency=chf --override checkout_session:metadata.id_dossier=1