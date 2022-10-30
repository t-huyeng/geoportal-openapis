import requests
import json
import sys
import time
from os.path import exists

main_url = ""


def call_api(i):
    url = main_url + str(i) + "/api"
    response = requests.get(url)
    return response.text


def check_api(i):
    text = call_api(i)
    if text.startswith("{") and text.endswith("}"):
        return True, text
    else:
        return False, ""


def call_api_all(folder, mappingfile):
    if not exists(mappingfile):
        print("no mapping file found")
        # create mapping file
        with open(mappingfile, "w") as f:
            f.write("[]")
    # read out mapping file
    mapping = []
    with open(mappingfile, "r") as f:
        mapping = json.load(f)
        for i in range(1, 1000):
            api, text = check_api(i)
            if api:
                # convert text to json and save title and i to textfile
                repoData = json.loads(text)
                title = repoData["info"]["title"]
                print(title)
                # check if api is already in mapping
                if (next((item for item in mapping if item["id"] == i), None)) is None:
                    # check if CORS is enabled for this api
                    url = main_url + str(i)
                    rawOpenAPI = url + "/api"
                    entry = {
                        "name": repoData["info"]["title"],
                        "url": url,
                        "rawOpenAPI": rawOpenAPI,
                        "id": i,
                    }
                    mapping.append(entry)
            else:
                print(i, "not found")
            time.sleep(0.1)
    with open(mappingfile, "w") as f:
        json.dump(
            mapping,
            f,
            indent=4,
        )


if __name__ == "__main__":
    main_url = "https://www.geoportal.rlp.de/spatial-objects/"
    call_api_all("geoportal-rlp", "geoportal_rlp_all.json")

    main_url = "https://www.geoportal.hessen.de/spatial-objects/"
    call_api_all("geoportal-he", "geoportal_he_all.json")

    main_url = "https://www.geoportal.saarland.de/spatial-objects/"
    call_api_all("geoportal-sl", "geoportal_sl_all.json")
    print("done")
    sys.exit(0)
