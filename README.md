# TP0: Docker + Comunicaciones + Concurrencia
- Alumna: Rocío Platini
- Padrón: 107456

## Instrucciones de ejecución
### Ejercicios 6, 7 y 8
Antes de levantar los contenedores se debe ejecutar el script `./generar-compose.sh docker-compose-dev.yaml 5` para crear a los 5 clientes, ya que no modifiqué las configuraciones provistas inicialmente por la cátedra.

### Ejercicios restantes
Se deben ejecutar de la forma indicada por la cátedra.

## Protocolo de comunicación
Decidí implementar un protocolo donde los bloques sean de tamaño variable, y los mensajes se serializan como un arreglo de bytes. 
Cuando se serializa un mensaje, se incluye en el paquete un *header*, que indica el tamaño del *payload* y además contiene un flag indicando si es el paquete final. El tamaño del header es fijo (5 bytes) donde los primeros 4 bytes corresponden al tamaño y el último es el flag. 

Cuando el receptor decodifica el mensaje, lee los primeros 5 bytes y puede obtener la cantidad exacta de *payload* que debe leer, además de saber si es el último paquete o debe esperar más.

Para serializar las apuestas, decidí utilizar caracteres delimitadores para separar la información. Por ejemplo, suponiendo que se quiere enviar el *batch* con las siguientes apuestas:

- `Santiago Lionel,Lorca,30904465,1999-03-17,2201`
- `Agustin Emanuel,Zambrano,21689196,2000-05-10,9325`

El *payload* del paquete se codificará de la siguiente forma:

`Santiago Lionel,Lorca,30904465,1999-03-17,2201\nAgustin Emanuel,Zambrano,21689196,2000-05-10,9325`

De esta forma, el servidor puede deserializar el mensaje splitteando por el caracter delimitador `\n` que indica la finalización de una apuesta, y obtener cada campo (nombre, apellido, etc.) separando los strings por una coma `,`.

Para consultar por los ganadores, los clientes envían el mensaje `WINNERS,1`, donde el número indica la agencia que está consultando. El servidor responderá un mensaje retornando los documentos de los ganadores, utilizando un caracter delimitador al igual que con las apuestas. 

En mi implementación decidí que cuando el servidor recibe exitosamente un *batch*, este no responda al cliente con ningún mensaje de ACK. Esta decisión la tomé basándome en que se está utilizando el protocolo de transporte TCP, que garantiza el correcto envío y recepción de paquetes.

## Parte 3
Para que el server pueda procesar mensajes en paralelo, decidí utilizar la librería de Python `multiprocessing`, de forma que por cada cliente que se conecta, el servidor crea un nuevo proceso para procesar y guardar los *batches* de datos.

Por cada nueva conexión entrante, el servidor almacena el socket junto con la dirección IP del cliente, y la lotería comienza si la cantidad de conexiones es igual a la cantidad de agencias que se espera recibir.

### Mecanismos de sincronización utilizados 
En cuanto a la sincronización, el único recurso compartido entre procesos es el archivo `bets.csv`, donde se guardan las apuestas recibidas por los clientes. Opté por la utilizaciòn de un Lock (provisto por la librerìa `multiprocessing`) para que solo un proceso por vez pueda escribir en el archivo.
