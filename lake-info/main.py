from lake_info import fetch_lake_data_table
from send_data import send_data

from os import getenv


def main():
    url = getenv("USACE_URL")
    prefix = getenv("LAKE_PREFIX")
    print("fetching {}".format(url))
    dt = fetch_lake_data_table(url=url)
    send_data(dt, prefix)


if __name__ == "__main__":
    main()
