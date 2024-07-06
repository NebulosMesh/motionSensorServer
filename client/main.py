import requests
import queue
import threading
import socket
import time
from gpiozero import MotionSensor


class NetworkBridge:
    def __init__(self):
        self.__socket_server_addr = ("192.168.10.188", 3333)
        self.__max_retries_per_chunk = 5
        self.__max_retry_chunks = 5
        self.__delay_multiplier_between_chunks = 2
        self.__max_retry_chunks = 5
        self.__currnet_attempts = 0
        self.__current_timeout_between_retries_in_seconds = 0.2
        self.__sender_version = 1
        self.__DEVICE_ID = 4
        self.__SOFTWARE_ID = 4
        self.__ID_Payload = bytearray([self.__DEVICE_ID])
        self.paired = False
        self.__bufferSize = 4
        self.__socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

    def send_data_to_endpoint(self, data):
        status = self.recursively_send_data_to_endpoint(data)
        if status == 'err':
            self.paired = False
            print("failed to send data")
            self.__currnet_attempts = 0
            self.pair_sensor()

    def recursively_send_data_to_endpoint(self, data):
        try:
            # self.post_data(json)
            self.socket_sender(data)
            return 'success'
        except:
            current_cunk = math.ceil(
                self.__currnet_attempts / self.__max_retries_per_chunk)
            current_pos_in_chunk = math.fmod(self.__currnet_attempts,
                                             math.floor(
                                                 self.__currnet_attempts /
                                                 self.__max_retries_per_chunk))

            if current_cunk == self.__max_retry_chunks and current_pos_in_chunk == self.__max_retries_per_chunk:
                return 'err'

            if current_pos_in_chunk == self.__max_retries_per_chunk:
                time.sleep(self.__current_timeout_between_retries)
                self.__current_timeout_between_retries = self.__current_timeout_between_retries * \
                    self.__delay_multiplier_between_chunks

            print("failed to send data, retrying...")
            self.__currnet_attempts += 1
            self.send_data_to_endpoint(data)

            if not self.__currnet_attempts >= self.__max_retries:
                print("failed to send data, retrying...")
                self.__currnet_attempts += 1
                self.send_data_to_endpoint(data)
            else:
                return 'err'

    def socket_sender(self, data):
        payload = bytearray([self.__sender_version]) + data
        fullPayload = bytearray([len(payload)]) + payload
        print(fullPayload)
        self.__socket.sendto(fullPayload, self.__socket_server_addr)

    def socket_reader(self):
        print("reading")
        message = self.__socket.recvfrom(self.__bufferSize)[0]
        if message[1] != 1:
            print("invalid message from server")
            return
        if message[2] == 0:
            print("error with paired with sensor")
            if self.paired:
                self.paired = False
        elif message[2] == 1:
            print("sensor is paired successfully")
            self.paired = True

    def pair_sensor(self):
        self.__ID_Payload = self.__ID_Payload + bytearray([self.__SOFTWARE_ID])
        while self.paired is False:
            self.send_data_to_endpoint(self.__ID_Payload)
            time.sleep(1)

    def socket_reader_daemon(self, _, stop_event):
        while not stop_event.is_set():
            self.socket_reader()
            time.sleep(1)


class Queue(object):
    motion_queue = queue.Queue(20)

    def __new__(cls):
        if not hasattr(cls, 'instance'):
            cls.instance = super(Queue, cls).__new__(cls)
        return cls.instance

    def add_item_to_queue(self, item):
        self.motion_queue.put(item)

    def get_item_in_queue(self):
        item = self.motion_queue.get()
        return item

    def flush_queue(self):
        with self.motion_queue.mutex:
            self.motion_queue.queue.clear()

    def queue_daemon(self, network, stop_event):
        network = network
        while not stop_event.is_set():
            try:
                work = self.get_item_in_queue()
            except queue.Empty:
                return
            network.send_data_to_endpoint(work)
            self.motion_queue.task_done()


class PidSensor:
    __OUTPUT_SIGNAL_BUFFER = 2.5
    __TIMEOUT_IN_SECONDS = __OUTPUT_SIGNAL_BUFFER
    __GPIO_PIN_NUMBER = 4

    def __init__(self, device_id, motion_queue):
        self.sensor = MotionSensor(self.__GPIO_PIN_NUMBER)
        self.__QUEUE = motion_queue
        self.device_id = device_id

    def run_sensor_lifecycle(self, payload):
        self.sensor.wait_for_motion()
        self.send_motion_triggered_event(payload)
        self.sensor.wait_for_no_motion(self.__TIMEOUT_IN_SECONDS)

    def send_motion_triggered_event(self, payload):
        print("movement detected")
        self.__QUEUE.add_item_to_queue(payload)


class Orchistrator:
    __DEVICE_ID = 4
    __SOFTWARE_ID = 4
    stop_event = threading.Event()

    def __init__(self):
        print("initializing eye")
        self.payload = bytearray([self.__DEVICE_ID])
        self.motion_queue = Queue()
        self.network_bridge = NetworkBridge()
        self.pid_sendor = PidSensor(self.__DEVICE_ID, self.motion_queue)
        self.start_threads()

        print("pairing eye")
        self.network_bridge.pair_sensor()
        print("the eye is ready")

    def start_threads(self):
        print("starting threads")
        self.queue_daemon = threading.Thread(
            target=self.motion_queue.queue_daemon,
            daemon=True, name="queueDaemon",
            args=(self.network_bridge, self.stop_event))
        self.socket_reader_daemon = threading.Thread(
            target=self.network_bridge.socket_reader_daemon,
            daemon=True, name="socketReader",
            args=("", self.stop_event))
        self.queue_daemon.start()
        self.socket_reader_daemon.start()

    def run(self):
        try:
            while True:
                if self.network_bridge.paired:
                    self.pid_sendor.run_sensor_lifecycle(self.payload)
                else:
                    self.network_bridge.pair_sensor()
        except KeyboardInterrupt:
            self.terminate()
            pass

    def terminate(self):
        print("closing the eye")
        self.motion_queue.flush_queue()
        self.stop_event.set()
        print("eye closed")


eye = Orchistrator()
eye.run()
