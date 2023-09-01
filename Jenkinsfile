pipeline {
  agent { label 'home' }
    environment {
        IMAGE           = 'dzailz/sr-api:latest'
        CONTAINER_NAME  = 'sr-api'
    }

  stages {
    stage('prepare') {
    steps {
      sh '''
        docker_image_id=$(docker images -q "${IMAGE}")

        if [ -n "$docker_image_id" ]; then
            if [ "$(docker stop ${CONTAINER_NAME})" ]; then
                echo "container stopped"
            fi

            if [ "$(docker rmi "$docker_image_id" -f)" ]; then
                echo "image removed"
            fi
        fi
      '''
      }
    }

    stage('run') {
      steps {
        sh 'docker run -d --env-file /var/sr-api/.env.local --name "${CONTAINER_NAME}" -p 8787:8080 --rm "${IMAGE}"'
      }
    }

    stage('test') {
      steps {
        sh '''
        attempt_counter=0
        max_attempts=10

        until { curl --output /dev/null --silent --get --fail http://127.0.0.1:8787; } do
            if [ ${attempt_counter} -eq ${max_attempts} ]; then
              echo "Max attempts reached"
              exit 1
            fi

            printf '.'
            attempt_counter=$((attempt_counter + 1))
            sleep 1
        done
        '''
      }
    }

  }
}
