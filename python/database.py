from sqlalchemy import create_engine
#from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, declarative_base

SQLALCHEMY_DATABASE_URL = 'sqlite:///python/db/mercari.sqlite3'

# Engine: Manages the connection to the database
engine = create_engine( 
    SQLALCHEMY_DATABASE_URL, connect_args={'check_same_thread': False}
)
# Session: Manages database operations
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()
