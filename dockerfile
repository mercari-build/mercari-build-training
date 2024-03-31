# Use the Python 3.9 slim variant as the base image
FROM python:3.9

# Set the working directory inside the container to /app
WORKDIR /app

# Copy the contents of the 'db' directory from the host to the '/app/db' directory in the container
COPY db/ /app/db

# Copy the contents of the 'python' directory from the host to the '/app/python' directory in the container
COPY python/ /app/python

# Set the user context to 'trainee'
# USER trainee

# Print the version of Python when the container starts
CMD ["python", "-V"]

# Install Python dependencies from the requirements.txt file located in the '/app/python' directory
RUN pip3 install -r /app/python/requirements.txt

# Expose port 9000 to allow communication with the container
EXPOSE 9000

# Start the FastAPI application using uvicorn when the container starts
CMD ["uvicorn", "python.main:app", "--host", "0.0.0.0", "--port", "9000"]
