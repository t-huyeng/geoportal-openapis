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
                    entry = {
                        "name": repoData["info"]["title"],
                        "url": main_url + str(i),
                        "rawOpenAPI": "https://raw.githubusercontent.com/t-huyeng/geoportal-openapis/main/"
                        + folder
                        + "/"
                        + str(i)
                        + ".json",
                        "id": i,
                    }
                    mapping.append(entry)
                if exists(folder + "/" + str(i) + ".json"):
                    print("OpenAPI exists already.")
                else:
                    with open(folder + "/" + str(i) + ".json", "w") as myfile:
                        myfile.write(text)
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
    call_api_all("geoportal-rlp", "geoportal_rlp.json")

    main_url = "https://www.geoportal.hessen.de/spatial-objects/"
    call_api_all("geoportal-he", "geoportal_he.json")

    main_url = "https://www.geoportal.saarland.de/spatial-objects/"
    call_api_all("geoportal-sl", "geoportal_sl.json")
    print("done")
    sys.exit(0)
