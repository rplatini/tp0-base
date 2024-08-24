import sys
import yaml

FILENAME = 1
CLIENTS_NUM = 2

def main():
    print(f'Generating {sys.argv[CLIENTS_NUM]} clients for {sys.argv[FILENAME]}')

    with open(sys.argv[FILENAME], 'r') as file:
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
            ]
        }

    with open(sys.argv[FILENAME], 'w') as file:
        yaml.dump(docker_compose_yaml, file)

main()


