import requests
import urllib3
import re
from datetime import datetime, timedelta
urllib3.disable_warnings()

table_rock_url = 'https://www.swl-wc.usace.army.mil/pages/data/tabular/htm/tab7d.htm'


def check_valid_row(row):
    """Checks if a row is a valid data table row."""
    data = row.split()
    valid_length = len(data) == 8
    valid_date = False
    valid_measurements = False
    if valid_length:
        date_pattern = re.compile(r'\d{2}\w{3}\d{4}')
        valid_date = date_pattern.match(data[0])
        try:
            float(data[2])  # lake level
            float(data[3])  # tailwater level
            float(data[4])  # power generation (mwh)
            float(data[5])  # turbine release (cfs)
            float(data[6])  # spillway release (cfs)
            float(data[7])  # total release (cfs)
            valid_measurements = True
        except ValueError:
            pass

    return valid_length and valid_date and valid_measurements


def make_data_packet(row):
    """Creates an object with the data."""
    data = row.split()

    measurement_datetime = ""

    time_string = data[1]
    if time_string == "2400":
        measurement_datetime = datetime.strptime(
            data[0]+"0000", '%d%b%Y%H%M') + timedelta(hours=24)
    else:
        measurement_datetime = datetime.strptime(data[0]+data[1], '%d%b%Y%H%M')

    lake_level = float(data[2])
    tailwater_level = float(data[3])
    power_generation_mwh = float(data[4])
    turbine_release_cfs = float(data[5])
    spillway_release_cfs = float(data[6])
    total_release_cfs = float(data[7])

    packet = {
        'timestamp': measurement_datetime,
        'lake_level': lake_level,
        'tailwater_level': tailwater_level,
        'power_generation_mwh': power_generation_mwh,
        'turbine_release_cfs': turbine_release_cfs,
        'spillway_release_cfs': spillway_release_cfs,
        'total_release_cfs': total_release_cfs
    }
    return packet


def fetch_lake_data_table(url=table_rock_url):
    """Fetch the USACE Data table."""

    req = requests.get(url, verify=False)
    table = req.text.split("\n")
    packets = []
    for r in table:
        if check_valid_row(r):
            # print(r)
            packets.append(make_data_packet(r))
    return packets


fetch_lake_data_table()
