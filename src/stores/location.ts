import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';

import {
    type LocationInfoResponse,
    type LocationNewDisplayOrderRequest,
    Location
} from '@/models/location.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useLocationsStore = defineStore('locations', () => {
    const allLocations = ref<Location[]>([]);
    const allLocationsMap = ref<Record<string, Location>>({});
    const locationListStateInvalid = ref<boolean>(true);

    const allVisibleLocations = computed<Location[]>(() => {
        const visibleLocations: Location[] = [];

        for (const location of allLocations.value) {
            if (!location.hidden) {
                visibleLocations.push(location);
            }
        }

        return visibleLocations;
    });

    const allAvailableLocationsCount = computed<number>(() => allLocations.value.length);

    function loadLocationList(locations: Location[]): void {
        allLocations.value = locations;
        allLocationsMap.value = {};

        for (const location of locations) {
            allLocationsMap.value[location.id] = location;
        }
    }

    function addLocationToList(location: Location): void {
        allLocations.value.push(location);
        allLocationsMap.value[location.id] = location;
    }

    function updateLocationInList(currentLocation: Location): void {
        for (const [location, index] of itemAndIndex(allLocations.value)) {
            if (location.id === currentLocation.id) {
                allLocations.value.splice(index, 1, currentLocation);
                break;
            }
        }

        allLocationsMap.value[currentLocation.id] = currentLocation;
    }

    function updateLocationDisplayOrderInList({ from, to }: { from: number, to: number }): void {
        allLocations.value.splice(to, 0, allLocations.value.splice(from, 1)[0] as Location);
    }

    function updateLocationVisibilityInList({ location, hidden }: { location: Location, hidden: boolean }): void {
        if (allLocationsMap.value[location.id]) {
            allLocationsMap.value[location.id]!.hidden = hidden;
        }
    }

    function removeLocationFromList(currentLocation: Location): void {
        for (const [location, index] of itemAndIndex(allLocations.value)) {
            if (location.id === currentLocation.id) {
                allLocations.value.splice(index, 1);
                break;
            }
        }

        if (allLocationsMap.value[currentLocation.id]) {
            delete allLocationsMap.value[currentLocation.id];
        }
    }

    function updateLocationListInvalidState(invalidState: boolean): void {
        locationListStateInvalid.value = invalidState;
    }

    function resetLocations(): void {
        allLocations.value = [];
        allLocationsMap.value = {};
        locationListStateInvalid.value = true;
    }

    function loadAllLocations({ force }: { force?: boolean }): Promise<Location[]> {
        if (!force && !locationListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allLocations.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllLocations().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve location list' });
                    return;
                }

                if (locationListStateInvalid.value) {
                    updateLocationListInvalidState(false);
                }

                const locations = Location.ofMulti(data.result);

                if (force && data.result && isEquals(allLocations.value, locations)) {
                    reject({ message: 'Location list is up to date', isUpToDate: true });
                    return;
                }

                loadLocationList(locations);

                resolve(locations);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load location list', error);
                } else {
                    logger.error('failed to load location list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve location list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveLocation({ location, beforeResolve }: { location: Location, beforeResolve?: BeforeResolveFunction }): Promise<Location> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<LocationInfoResponse>;

            if (!location.id) {
                promise = services.addLocation(location.toCreateRequest());
            } else {
                promise = services.modifyLocation(location.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!location.id) {
                        reject({ message: 'Unable to add location' });
                    } else {
                        reject({ message: 'Unable to save location' });
                    }
                    return;
                }

                const newLocation = Location.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!location.id) {
                            addLocationToList(newLocation);
                        } else {
                            updateLocationInList(newLocation);
                        }
                    });
                } else {
                    if (!location.id) {
                        addLocationToList(newLocation);
                    } else {
                        updateLocationInList(newLocation);
                    }
                }

                resolve(newLocation);
            }).catch(error => {
                logger.error('failed to save location', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!location.id) {
                        reject({ message: 'Unable to add location' });
                    } else {
                        reject({ message: 'Unable to save location' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideLocation({ location, hidden }: { location: Location, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideLocation({
                id: location.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this location' });
                    } else {
                        reject({ message: 'Unable to unhide this location' });
                    }
                    return;
                }

                updateLocationVisibilityInList({ location, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change location visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this location' });
                    } else {
                        reject({ message: 'Unable to unhide this location' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeLocationDisplayOrder({ locationId, from, to }: { locationId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentLocation = allLocationsMap.value[locationId];

            if (!currentLocation || !allLocations.value[to]) {
                reject({ message: 'Unable to move location' });
                return;
            }

            if (!locationListStateInvalid.value) {
                updateLocationListInvalidState(true);
            }

            updateLocationDisplayOrderInList({ from, to });

            resolve();
        });
    }

    function updateLocationDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: LocationNewDisplayOrderRequest[] = [];

        for (const [location, index] of itemAndIndex(allLocations.value)) {
            newDisplayOrders.push({
                id: location.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveLocation({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move location' });
                    return;
                }

                if (locationListStateInvalid.value) {
                    updateLocationListInvalidState(false);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save locations display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move location' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteLocation({ location, beforeResolve }: { location: Location, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteLocation({
                id: location.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this location' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeLocationFromList(location);
                    });
                } else {
                    removeLocationFromList(location);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete location', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this location' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        allLocations,
        allLocationsMap,
        locationListStateInvalid,
        allVisibleLocations,
        allAvailableLocationsCount,
        updateLocationListInvalidState,
        resetLocations,
        loadAllLocations,
        saveLocation,
        hideLocation,
        changeLocationDisplayOrder,
        updateLocationDisplayOrders,
        deleteLocation
    }
});
