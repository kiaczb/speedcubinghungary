from database.db import upsert_person, get_person_data, get_all_important_delegates_from_db

difference = 3
milestones = [50, 100]

def GetImportantDelegates():
    return get_all_important_delegates_from_db(milestones, difference)

def AddCompetitionToDelegate(delegate):
    if not delegate.WcaId:
        return

    current = get_person_data(delegate.WcaId)
    new_count = (current['delegated_count'] + 1) if current and current.get('delegated_count') else 1

    upsert_person(delegate.WcaId, delegate.CompetitorName, del_count=new_count)

def IsHungarianDelegate(person):
    if not person.WcaId:
        return False

    db_person = get_person_data(person.WcaId)
    return bool(db_person and db_person.get('delegated_count', 0) > 0)
