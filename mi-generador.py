import sys
import yaml

FILENAME = 1
CLIENTS_NUM = 2

def main():
    fname = sys.argv[FILENAME]

    try:
        file = open(fname, 'r')

    except OSError:
        print(f'Could not open/read file: {fname}') 
        sys.exit(1)

    with file:
        docker_compose_yaml = yaml.safe_load(file)

        for i in range(2, int(sys.argv[CLIENTS_NUM]) + 1):
            client_name = f'client{i}'

            docker_compose_yaml['services'][client_name] = {
                'container_name': f'{client_name}',
                'image': 'client:latest',
                'entrypoint': '/client',
                'environment': [
                    f'CLI_ID={i}' ,
                    'CLI_LOG_LEVEL=DEBUG'
                ],
                'networks': [
                    'testing_net'
                ],
                'depends_on': [
                    'server'
                ],
                'volumes': [
                    './client/config.yaml:/config.yaml'
                ]
            }

            file.close()

    try:
        file = open(fname, 'w')
    
    except OSError:
        print(f'Could not open/read file: {fname}') 
        sys.exit(1)

    with file:
        yaml.dump(docker_compose_yaml, file)
        file.close()

main()


