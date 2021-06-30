from logging import exception
from requests_html import HTMLSession
import urllib3
import re
urllib3.disable_warnings()

table_rock_url = 'https://anglerspy.com/table-rock-lake-water-temperature-ipm/'


def fetch_lake_temp(url=table_rock_url):
    """Fetch the temperature."""
    try:
        session = HTMLSession()
        r = session.get(url, verify=False)

        temp_html = r.html.find("#temperature-fahrenheit", first=True)
        print(temp_html.raw_html)
        print(temp_html.text)
        lake_temp = float(temp_html.text[:-2])
        return lake_temp
    except Exception as e:
        print("Exception! {}".format(e))
        return None


if __name__ == "__main__":
    fetch_lake_temp()
