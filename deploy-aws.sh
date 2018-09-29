#!/usr/bin/env sh 

PROJECT_NAME=$1
SHA1=$2 # Nab the SHA1 of the desired build from a command-line argument
BRANCH=$3
EB_BUCKET=elasticbeanstalk-us-west-2-194925301021

# Create new Elastic Beanstalk version
echo $BRANCH
DOCKERRUN_FILE=$PROJECT_NAME-$BRANCH-bundle.zip

if [ "$BRANCH" == "master" ]; then 
    echo "yo"

    TAG=development
    ENV_NAME=$PROJECT_NAME-env

elif [ "$BRANCH" == "stage" ]; then

    TAG=stage
    ENV_NAME=$PROJECT_NAME-stg

elif [ "$BRANCH" == "production" ]; then

    TAG=latest
    ENV_NAME=$PROJECT_NAME-env
fi

echo $ENV_NAME
echo $DOCKERRUN_FILE
sed "s/<TAG>/$TAG/" < Dockerrun.aws.json > eb-source-bundle/Dockerrun.aws.json
cd ./eb-source-bundle/ ; zip -r ../$DOCKERRUN_FILE .ebextensions Dockerrun.aws.json; cd ..
aws configure set default.region us-west-2
aws configure set region us-west-2

aws s3 cp $DOCKERRUN_FILE s3://$EB_BUCKET/$DOCKERRUN_FILE # Copy the Dockerrun file to the S3 bucket
aws elasticbeanstalk create-application-version --application-name $PROJECT_NAME --version-label $SHA1 --source-bundle S3Bucket=$EB_BUCKET,S3Key=$DOCKERRUN_FILE

# Update Elastic Beanstalk environment to new version
aws elasticbeanstalk update-environment --environment-name $ENV_NAME --version-label $SHA1
