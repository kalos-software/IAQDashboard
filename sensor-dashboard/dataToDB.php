<?php

// variables for sesnor values from Arduino
$location = $_POST['location'];
$temp = $_POST['temp'];
$rH = $_POST['rH'];
$pmass1 = $_POST['pmass1'];
$pmass25 = $_POST['pmass25'];
$pmass4 = $_POST['pmass4'];
$pmass10 = $_POST['pmass10'];
$pcount1 = $_POST['pcount1'];
$pcount25 = $_POST['pcount25'];
$pcount4 = $_POST['pcount4'];
$pcount10 = $_POST['pcount10'];
$typPartSize = $_POST['typPartSize'];
$HCHO = $_POST['HCHO'];
$CO2 = $_POST['CO2'];

// variables for calculated metrics
$indoorTd = $temp-((100-$rH)/5);

// SQL server informaton
$host = "localhost";
$username = "data";
$password = "mysqldatapassword";
$dbname = "buildingData";

// create connection with database
$con = mysqli_connect($host, $username, $password, $dbname);

if (mysqli_connect_errno()) {
        die("connection error: " . mysqli_connect_error());
}

//insert sensor values into database
$sql = "INSERT INTO IAQ (location, temp, rH, pmass1, pmass25, pmass4, pmass10, pcount1, pcount25, pcount4, pcount10, typPartSize, HCHO, CO2, indoorTd)
               VALUES ('$location', '$temp', '$rH', '$pmass1', '$pmass25', '$pmass4', '$pmass10', '$pcount1', '$pcount25', '$pcount4', '$pcount10', '$typPartSize', '$HCHO', '$CO2', '$indoorTd')";

//$textFile = fopen('test.txt', 'w');
//fwrite($textFile, $sql);
//fclose($textFile);


mysqli_query($con, $sql);

mysqli_close($con);
