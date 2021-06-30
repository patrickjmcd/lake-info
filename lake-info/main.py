from lake_level import fetch_lake_data_table
from lake_temp_whiteriversky import fetch_lake_temp
from send_data import send_data

from os import getenv


def main():
    url = getenv("USACE_URL")
    prefix = getenv("LAKE_PREFIX")
    temperature_url = getenv("TEMPERATURE_URL")

    print("fetching {}".format(url))
    dt = fetch_lake_data_table(url=url)
    lake_temp = None
    if temperature_url:
        lake_temp = fetch_lake_temp(temperature_url)

    send_data(dt, prefix, lake_temp=lake_temp)


if __name__ == "__main__":
    main()
