from database.db import upsert_person, get_person_data, get_all_important_delegates_from_db

difference = 3
milestones = [50, 100]

def GetImportantDelegates(persons=None):
    # Teljesen függetlenítjük a kapott listától, a DB-ből szedjük ki az ÖSSZES embert
    return get_all_important_delegates_from_db(milestones, difference)

def AddCompetitionToDelegate(delegate):
    if not delegate.WcaId:
        return
    
    # Lekérjük a jelenlegi adatokat
    current = get_person_data(delegate.WcaId)
    
    # Ha van már adata és van delegated_count, akkor +1, különben 1 lesz az érték
    new_count = (current['delegated_count'] + 1) if current and current.get('delegated_count') else 1
    
    # Adatbázis frissítése / beszúrása
    upsert_person(delegate.WcaId, delegate.CompetitorName, del_count=new_count)

def IsImportantDelegate(wcaId):
    if not wcaId:
        return False
        
    person = get_person_data(wcaId)
    if not person:
        return False
    
    count = person.get('delegated_count', 0)
    if count is None:
        count = 0
        
    for milestone in milestones:
        if milestone - difference <= count <= milestone:
            return True
    return False

def IsHungarianDelegate(person):
    if not person.WcaId:
        return False
        
    db_person = get_person_data(person.WcaId)
    
    # A régi "benne van a JSON-ben" logikát úgy helyettesítjük, 
    # hogy megnézzük: létezik-e az adatbázisban, ÉS van-e delegált versenye.
    if db_person and db_person.get('delegated_count', 0) > 0:
        return True
        
    return False