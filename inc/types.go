package inc

/* cxxcam - C++ CAD/CAM driver library.
 * Copyright (C) 2013  Nicholas Gill
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
 * types.h
 *
 *  Created on: 2013-08-27
 *      Author: nicholas
 */

const (
	RS274NGC_TEXT_SIZE = 256
	// array sizes
	RS274NGC_ACTIVE_G_CODES  = 12
	RS274NGC_ACTIVE_M_CODES  = 7
	RS274NGC_ACTIVE_SETTINGS = 3
	// number of parameters in parameter table
	RS274NGC_MAX_PARAMETERS = 5400
)

const (
	TOLERANCE_INCH float64 = 0.0002
	TOLERANCE_MM   float64 = 0.002
	/* angle threshold for concavity for cutter compensation, in radians */
	TOLERANCE_CONCAVE_CORNER float64 = 0.01
)

/* numerical constants */
const (
	TINY    float64 = 1e-12 /* for arc_data_r */
	UNKNOWN         = 1e-20
	TWO_PI          = 6.2831853071795864
	PI              = 3.1415926535897932
	PI2             = 1.5707963267948966
)

// English - Metric conversion (long number keeps error buildup down)
const (
	MM_PER_INCH = 25.4
	INCH_PER_MM = 0.039370078740157477
)

type Units int

const (
	_ Units = iota
	Imperial
	Metric
)

// feed_mode
type FeedMode int

const (
	_ FeedMode = iota
	UNITS_PER_MINUTE
	INVERSE_TIME
	//UnitsPerMinute
	//InverseTime
)

type SpindleMode int

const (
	_ SpindleMode = iota
	ConstantRPM
	ConstantSurface
)

//type DistanceMode int
//const (
//	_ DistanceMode = iota
//	Absolute
//	Incremental
//)
/* distance_mode */
type DISTANCE_MODE int

const (
	_ DISTANCE_MODE = iota
	MODE_ABSOLUTE
	MODE_INCREMENTAL
)

/* retract_mode for cycles */

type RETRACT_MODE int

const (
	_ RETRACT_MODE = iota
	R_PLANE
	OLD_Z
)
