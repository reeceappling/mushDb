import Box from '@mui/material/Box';
import Slider from '@mui/material/Slider';
import {useState} from "react";

export function WetnessSliderInternal({defaultValue,onChange}:{
    defaultValue:number,
    onChange: (event: Event, value: number, activeThumb: number) => void,
}) {
    return (
        <Box sx={{width: 300}}> {/* TODO: fix box size*/}
            <Slider
                min={0} max={10} defaultValue={defaultValue} step={1}
                size="medium" // small, medium, large
                aria-label="Wetness" // Label

                valueLabelDisplay="on" /* Can be off, on, or auto */
                marks={[{value: 0,label: "very dry"},
                    {value: 5,label: "perfect"},
                    {value: 10,label: "very wet"},
                ]}
                getAriaValueText={(
                    value: number,
                    index: number,
                )=>{return value.toString()}}
                /*color={'primary'} TODO: ???*/
                onChange={onChange}
            />
        </Box>
        // <div>
        //     <div className={"inline"}>{"Wetness"}</div>
        //     <div className={"inline"}>
        //
        //     </div>
        // </div>
    );
}

export default function WetnessSlider({defaultValue,onChange}:
                                      {
                                          defaultValue:number,
                                          onChange: (event: Event, value: number, activeThumb: number) => void,
                                      }) {
    return (
        <div className={"wetnessSliderContainer"}>
            <div className={"wetnessLabel"}>{"Wetness: "}</div>{/* TODO: LABEL!*/}
            <div className={"wetnessSlider"}>
                <WetnessSliderInternal defaultValue={defaultValue} onChange={onChange}/>
            </div>
        </div>

    );
}
