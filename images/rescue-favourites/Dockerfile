FROM python:3.15-alpine@sha256:a943ed02f4e22d024d2b4fc5083619497dfb9bc1c64c193861f3bc6d3074f5b1
ENV PYTHONDONTWRITEBYTECODE=1
COPY rescue_favourites.py /app/rescue_favourites.py
USER 1000:1000
ENTRYPOINT ["python", "-I", "/app/rescue_favourites.py"]
