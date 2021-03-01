FROM python:3

WORKDIR /usr/src/app

COPY requirements.txt ./

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# + work init but not flask
#ENV FLASK_APP "./__init__.py"
ENV FLASK_APP "./app.py"


EXPOSE 5000

# Works but without  gen
#CMD [ "python", "./pdf_generator.py" ]
#CMD [ "python", "-u", "./app.py" ]
# Same
 CMD python ./pdf_generator.py
 CMD [ "python", "-m", "flask", "run", "--host", "0.0.0.0" ]


# No module named 'config'
#CMD flask run --host=0.0.0.0


# "--host", "0.0.0.0" -m

# Ни __init__ ни порты

# Generator works
#CMD [ "python", "-m", "flask" ]

# No port, no generator


#CMD python -m flask run