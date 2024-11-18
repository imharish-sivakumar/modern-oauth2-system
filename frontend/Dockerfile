# Use official Node.js image as the base image for building the app
FROM node:18 AS build

# Set the working directory
WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm install

# Copy the rest of the application files
COPY . .
COPY .env .env

# Build the React application
RUN npm run build

# Use Node.js to serve the React build files with `serve`
FROM node:18-alpine

# Install `serve` globally
RUN npm install -g serve

# Set the working directory
WORKDIR /app

# Copy only the built React app from the previous stage
COPY --from=build /app/build /app/build

# Expose port 5000 (or any port you prefer)
EXPOSE 80

# Start the server using `serve`
CMD ["serve", "-s", "build", "-l", "80"]
