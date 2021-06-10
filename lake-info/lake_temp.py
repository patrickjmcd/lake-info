from requests_html import HTMLSession
import urllib3
import re
from datetime import datetime, timedelta
urllib3.disable_warnings()

table_rock_url = 'https://anglerspy.com/table-rock-lake-water-temperature-ipm/'

def fetch_lake_temp(url=table_rock_url):
    """Fetch the temperature."""
    session = HTMLSession()
    r = session.get(url, verify=False)

    temp_html = r.html.find("#temp")[0]
    lake_temp = float(temp_html.text[:-2])
    return lake_temp

if __name__ == "__main__":
    fetch_lake_temp()
