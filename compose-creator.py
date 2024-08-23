import yaml
import argparse

def generate_client(index):
    return {
        f'client{index}': {
            'container_name': f'client{index}',
            'image': 'client:latest',
            'entrypoint': '/client',
            'environment': [
                f'CLI_ID={index}',
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
    }

def main():
    parser = argparse.ArgumentParser(description="Add services to docker-compose-dev.yaml")
    parser.add_argument('file', type=str, help='Output file')
    parser.add_argument('x', type=int, help='The number of services to add')
    args = parser.parse_args()

    base_file = 'docker-compose-base.yaml'
    output_file = args.file
    with open(base_file, 'r') as file:
        base_content = yaml.safe_load(file)

    if 'services' not in base_content:
        base_content['services'] = {}

    for i in range(1, args.x + 1):
        client = generate_client(i)
        base_content['services'].update(client)

    with open(output_file, 'w') as file:
        yaml.safe_dump(base_content, file, default_flow_style=False)

if __name__ == '__main__':
    main()