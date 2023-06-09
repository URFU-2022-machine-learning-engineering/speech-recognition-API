pipeline {
  agent { label 'home' }
  stages {
    stage('prepare') {
    steps {
      sh '''
        image="dzailz/sr-api:latest"
        container_name="sr-api"
        docker_image_id=$(docker images -q "$image")

        if [ ! -z "$docker_image_id" ]; then
            if [ $(docker stop "$container_name") ]; then
                echo "container stopped"
            fi
            if [ $(docker rmi "$docker_image_id" -f) ]; then
                echo "image removed"
            fi
        fi

      '''
      }
    }

    stage('run') {
      steps {
        sh 'docker run -d --env-file /var/sr-api/.env.local --name sr-api -p 8787:8080 --rm "$image"'
      }
    }

    stage('test') {
      steps {
        sh '''
        attempt_counter=0
        max_attempts=10

        until $(curl --output /dev/null --silent --get --fail http://192.168.111.66:8787); do
            if [ ${attempt_counter} -eq ${max_attempts} ];then
              echo "Max attempts reached"
              exit 1
            fi

            printf '.'
            attempt_counter=$(($attempt_counter+1))
            sleep 1
        done

        '''
      }
    }

  }
}
