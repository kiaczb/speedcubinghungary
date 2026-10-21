from datetime import datetime
def FormatSeconds_mm_ss(seconds):
    minutes = seconds // 60
    seconds = seconds % 60
    return f"{int(minutes)}:{round(seconds,2):.2f}"

def ConvertResult(time, category):

    if category == "333fm" and time < 100:
        return time
    if time >= 100000:
        missed = time % 100
        formattedTime = FormatSeconds_mm_ss(int(str(time)[2:7])).split(".")[0]
        point = 99-int(str(time)[:2])
        return f"{missed+point} / {2*missed+point} {formattedTime}"
    if time < 6000:
        formattedTime = f"{time/100:.2f}"
    else:
        formattedTime = f"{FormatSeconds_mm_ss(time/100)}"
    return formattedTime

def ConvertDate(fromDate, toDate):
    if fromDate == toDate:
        return f"{fromDate.strftime('%Y. %b. %d.')}"
    return f"{datetime.strftime(fromDate,'%Y. %b. %d-')}{datetime.strftime(toDate,'%d.')}"


def FlagEmoji(country_iso):
    if not country_iso or len(country_iso) != 2:
        return country_iso
    return "".join(chr(127397 + ord(c.upper())) for c in country_iso)

def ReverseHungarianName(name):
    if not name:
        return name
    parts = name.split()
    if len(parts) < 2:
        return name

    return f"{parts[-1]} {' '.join(parts[:-1])}"