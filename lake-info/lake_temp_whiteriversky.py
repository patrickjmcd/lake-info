import requests


table_rock_url = "https://api.whiteriversky.net/v1/temperatures/sites/5/"


def fetch_lake_temp(url=table_rock_url):
    """Fetch the temperature."""
    try:
        headers = {
            "Authorization": "Token 309d1fdcd670707df742ba1231c84ec3cbaf7041"
        }
        r = requests.get(url, headers=headers)
        temp_fahrenheit = float(r.json()['fahrenheit'])
        return temp_fahrenheit
    except Exception as e:
        print("Exception! {}".format(e))
        return None


if __name__ == "__main__":
    print(fetch_lake_temp())
