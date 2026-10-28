import logging
import requests
from database.db import upsert_record, get_all_records

logger = logging.getLogger(__name__)

localNationalRecords = get_all_records()

def SaveRecords():
    recordsUrl = "https://www.worldcubeassociation.org/api/v0/records"
    response = requests.get(recordsUrl, timeout=10)
    if response.status_code != 200:
        logger.error("HTTP %s while fetching records, records not updated", response.status_code)
        return

    recordsData = response.json()

    regions = {
        "world": recordsData["world_records"],
        "europe": recordsData["continental_records"].get("_Europe", {}),
        "hungary": recordsData["national_records"].get("Hungary", {})
    }

    for region_db_name, region_data in regions.items():
        for event_id, times in region_data.items():
            single = times.get('single')
            average = times.get('average')
            upsert_record(region_db_name, event_id, single, average)

    logger.info("Records saved.")
