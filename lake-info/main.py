from lake_level import fetch_lake_data_table
from lake_temp import fetch_lake_temp
from send_data import send_data

from os import getenv


def main():
    url = getenv("USACE_URL")
    prefix = getenv("LAKE_PREFIX")
    level_url = getenv("ANGLERSPY_URL")

    print("fetching {}".format(url))
    dt = fetch_lake_data_table(url=url)
    lake_level = None
    if level_url:
        lake_level = fetch_lake_temp(level_url)

    send_data(dt, prefix, level_value=lake_level)


if __name__ == "__main__":
    main()
