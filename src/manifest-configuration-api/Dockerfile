FROM mcr.microsoft.com/dotnet/sdk:5.0 AS build-env
WORKDIR /app

COPY *.csproj ./
RUN dotnet restore

COPY . ./
RUN dotnet publish -c Release -o out

FROM mcr.microsoft.com/dotnet/sdk:5.0
WORKDIR /app
COPY --from=build-env /app/out .

ENV ASPNETCORE_URLS="http://0.0.0.0:5000"
EXPOSE 5000
ENTRYPOINT ["dotnet", "ota-update-management.dll"]