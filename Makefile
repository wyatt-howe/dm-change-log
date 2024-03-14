AWS_REGION=us-east-1
REGISTRY=745821988602.dkr.ecr.${AWS_REGION}.amazonaws.com
BASE_IMAGE=dm-change-log

aws_login:
	aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${REGISTRY}

stack:
	clear && docker compose down -v && docker compose build && docker compose up --remove-orphans