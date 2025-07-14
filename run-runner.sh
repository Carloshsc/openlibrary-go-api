kubectl run runner-$(date +%s) \
  --restart=Never \
  --rm -it \
  --image=book-runner:latest \
  --image-pull-policy=Never \
  --env=API_HOST=book-api \
  --env=API_PORT=3000 \
  -- "$@" 2>&1 | grep -vE 'If you don|pod "[^"]+" deleted'

