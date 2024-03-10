pipeline {
    // Using any agent, as we're working on single instance of Jenkins
    agent any
    // Define environment variables
    environment {
        // Docker
        IMAGE_NAME = "scale-schedule"
        DOCKER_USERNAME = "ahmed3sam"
        DOCKER_PASSWORD = credentials('dockerhub_pass')

    }

    // Begin stages
     stages {

        //Build
        stage('Docker Build') {
            steps {
                    sh "docker build -t $IMAGE_NAME ."
                    sh "docker tag $IMAGE_NAME:latest $DOCKER_USERNAME/$IMAGE_NAME:latest"
            }
        }

        //Test
        stage('Docker Test') {
            agent {
                docker { 
                    image 'aquasec/trivy:latest'
                    args '--entrypoint=""'
                    reuseNode true   
                }
            }
            environment {
                HOME = "."
            }
            steps {
                    sh "trivy image --format template --template \"@/contrib/junit.tpl\" -o junit-report.xml  $DOCKER_USERNAME/$IMAGE_NAME:latest"
                    junit skipPublishingChecks: true, testResults: 'junit-report.xml'
            }
        }


        stage('Docker Push') {
            steps {
                    sh "docker login -u=\"$DOCKER_USERNAME\" -p=\"$DOCKER_PASSWORD\""
                    sh "docker push $DOCKER_USERNAME/$IMAGE_NAME:latest"
            }
        }
    }
    post {
        always {
            deleteDir()
        }
    }
}
