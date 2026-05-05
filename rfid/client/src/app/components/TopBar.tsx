"use client"

import ReaderWriterSelector, {
    ReadTagFunc,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import * as React from "react";
import {useState} from "react";
import Button from "@mui/material/Button"
import Menu from "@mui/material/Menu"
import MenuItem from "@mui/material/MenuItem"
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getTypeFor} from "@/app/components/common";
import {BaseExternalUrl} from "@/app/components/Constants";
import {TailwindButton} from "@/app/components/tailwind/components";


const buttonProps = {
    backgroundColor: 'var(--topBarColor)',
    color: 'var(--topBarTextColor)',
    '&:hover': {
        color: 'white',
        backgroundColor: 'var(--topBarHoverColor)',
    },
    '&:active': {
        color: 'white',
        backgroundColor: 'var(--topBarActiveColor)',
    }
}
const sublistItemProps = {
    // backgroundColor: 'var(--topBarSubmenuColor)',
    color: 'var(--topBarSubmenuTextColor)',
    '&:hover': {
        backgroundColor: 'var(--topBarSubmenuHoverColor)',
    },
    '&:active': {
        backgroundColor: 'var(--topBarSubmenuClickColor)',
    }
}

export function TopBarCreateMenu() {
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl)
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorEl(event.currentTarget)
    }
    const handleClose = () => {
        setAnchorEl(null)
    }
    return <div>
        <Button
            id={"topBarCreateButton"}
            sx={buttonProps}
            aria-controls={open ? 'topBarCreateMenu' : undefined}
            aria-haspopup={true}
            aria-expanded={open ? 'true' : undefined}
            onClick={handleClick}>
            {"Create"}
        </Button>
        <Menu id={"topBarCreateMenu"}
              anchorEl={anchorEl}
              open={open}
              onClose={handleClose}
              slotProps={{
                  list: {'aria-labelledby': 'topBarCreateButton'}
              }}>
            <MenuItem href={"/new/agarRecipe"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Agar Recipe"}</MenuItem>
            <MenuItem href={"/new/jarRecipe"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Grain Jar Recipe"}</MenuItem>
            <MenuItem href={"/new/lcRecipe"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Liquid Culture Recipe"}</MenuItem>
            <MenuItem href={"/new/project"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Project"}</MenuItem>{/* TODO: maybe just create this in each form? */}
            {/* TODO: PC RUN??? */}
            <MenuItem href={"/new/species"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Species"}</MenuItem>
            <MenuItem href={"/new/substrateRecipe"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Substrate Recipe"}</MenuItem>
            <MenuItem href={"/new/waterJar"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Water Jar"}</MenuItem> {/* TODO: ALSO ALLOW IT TO BE DONE FROM THE PC RUN PAGE*/}
        </Menu>
    </div>
}

export default function TopBar() {
    // TODO: RECENTS FOR ALL ENTRIES?????
    const {dispatch} = useRfidReaderContext()
    // TODO: ENSURE THE NEXT FUNCTION IS USED IN MULTIPLE PLACES!
    const onReaderSelect = (s: string | undefined) => {
        let session = "" // TODO: fix session
        ReadTagFunc(dispatch, session, s).then(id=>{
        // todo: do nothing with id result
        },err=>{
            console.error(err) // TODO: ok?
        })
    }
    return <div id={"topBar"}>
        <TopBarListMenu/>
        <TopBarViewMenu/>
        <TopBarImportMenu/>
        <TopBarCreateMenu/>
        <div id={"rfidTopArea"}>
            <LastReadTag/>
            <ReadTagButton/>
            <ReaderWriterSelector onSelect={onReaderSelect}/>
        </div>
    </div>
}

function CopyLatestReadTagButton() {
    const {state, dispatch} = useRfidReaderContext()
    if (state.lastReadTag == undefined) {
        return null
    }
    const onClick = () => {
        if (state.lastReadTag != undefined) {
            navigator.clipboard.writeText(state.lastReadTag).catch((err) => {
                let toWrite = "failed to copy tag value to clipboard: " + err
                console.error(toWrite)
                dispatch({
                    type: ActionTypes.SET_ERROR,
                    payload: toWrite,
                })
            })
        }
    }
    return <button className={"basicButtonSmall"} onClick={onClick}>{"Copy last read tag value"}</button>
}

function LastReadTag() {
    const {state} = useRfidReaderContext()
    if (state.lastReadTag !== undefined) {
        return <div>
            <div className={"centerH"}>{"Last read tag value: "}</div>
            <div className={"centerH"}>{state.lastReadTag}
                <button className={"basicButtonSmall"} onClick={() => {
                    state.lastReadTag && navigator.clipboard.writeText(state.lastReadTag)
                }}>{"Copy"}</button>
            </div>
        </div>
    }
    return <div>
        {"No tag value read yet"}
    </div>
}

export function Makeid(length: number) { // TODO: DELETEME
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    let counter = 0;
    while (counter < length) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
        counter += 1;
    }
    return result;
}

function UseLatestReadTagButton({onClick}: { onClick: (id?: string) => void }) {
    const {state, dispatch} = useRfidReaderContext()
    const onButtonClick = () => {
        onClick(state.lastReadTag)
    }
    return <button onClick={onButtonClick}>{"Use latest read tag id"}</button>
}


function ReadTagButton({onResult}: { onResult?: (id: string) => void }) {
    const {state, dispatch} = useRfidReaderContext()
    const onClick = () => {
        if (state.selected != undefined) {
            //ReadRfidTag(state.selected) // TODO: REENABLE
            const a = new Promise<string>((accept) => {// TODO: DELETE
                accept(Makeid(5))
            })
            a.then((tagVal) => {
                onResult && onResult(tagVal)
                dispatch({
                    type: ActionTypes.SET_LAST_READ_TAG,
                    payload: tagVal,
                })
                dispatch({
                    type: ActionTypes.SET_LAST_READER,
                    payload: state.selected,
                })
            }, (err) => {
                let toWrite = "failed to read tag: " + err
                console.error(toWrite)
                dispatch({
                    type: ActionTypes.SET_ERROR,
                    payload: toWrite,
                })
            })
        } else {
            let toWrite = "cannot read tag without knowing which reader to use!"
            console.error(toWrite)
            dispatch({
                type: ActionTypes.SET_ERROR,
                payload: toWrite,
            })
        }

    }
    return <button className={"basicButtonSmall"} onClick={e=>{
        e.stopPropagation();
        onClick();
    }}>{"Read Tag"}</button>
}

export function TopBarViewMenu() {
    const [id, setId] = useState("")
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl)
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorEl(event.currentTarget)
    }
    const handleClose = () => {
        setAnchorEl(null)
    }
    const handleViewById = () => {
        getTypeFor(id).then((entryType) => {
            location.assign(BaseExternalUrl + "/view/" + entryType + "/" + id)
        }).catch((err) => {
            // TODO: handle the error!
            console.log(err)
        })
    }
    return <div>
        <Button
            id={"topBarViewButton"}
            sx={buttonProps}
            aria-controls={open ? 'topBarViewMenu' : undefined}
            aria-haspopup={true}
            aria-expanded={open ? 'true' : undefined}
            onClick={handleClick}>
            {"View"}
        </Button>
        <Menu id={"topBarViewMenu"}
              anchorEl={anchorEl}
              open={open}
              onClose={handleClose}
              slotProps={{
                  list: {'aria-labelledby': 'topBarViewButton'}
              }}>
            <MenuItem onClick={() => {
            }}>
                <div>
                    <div>{"Main Collection Item By ID"}</div>
                    <TextBox readonly={false} label={"ID"} value={id} fieldName={"viewByIdInput"}
                             updateTextHandler={setId}/>
                    <ReadTagButton onResult={setId}/>
                    <UseLatestReadTagButton onClick={(v) => {
                        v && setId(v)
                    }}/>
                    <button onClick={handleViewById}> {"go to this id"}</button>
                </div>
            </MenuItem>
            {/* TODO: the rest*/}
            <MenuItem href={"/view/agarBatch"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Agar Batch"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/agarRecipe"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Agar Recipe"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/jarRecipe"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Jar Recipe"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/lcRecipe"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Liquid Culture Recipe"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/pcRun"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"PC Run"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/project"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Project"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/sale"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Sale"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/species"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Species"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/subspecies"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Subspecies"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/substrateBatch"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Substrate Batch"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/substrateRecipe"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Substrate Recipe"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/transfer"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Transfer"}</MenuItem>{/* TODO: FIX */}
            <MenuItem href={"/view/user"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"User"}</MenuItem>{/* TODO: FIX */}
            {/* TODO: INPUT OR SCAN FOR ITEM!*/}
        </Menu>
    </div>
}

export function TopBarImportMenu() {
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl)
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorEl(event.currentTarget)
    }
    const handleClose = () => {
        setAnchorEl(null)
    }
    return <div>
        <Button
            id={"topBarImportButton"}
            sx={buttonProps}
            aria-controls={open ? 'topBarImportMenu' : undefined}
            aria-haspopup={true}
            aria-expanded={open ? 'true' : undefined}
            onClick={handleClick}>
            {"Import"}
        </Button>
        <Menu id={"topBarImportMenu"}
              anchorEl={anchorEl}
              open={open}
              onClose={handleClose}
              slotProps={{
                  list: {'aria-labelledby': 'topBarImportButton'}
              }}>
            <MenuItem href={"/import/bag"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Bag"}</MenuItem>
            <MenuItem href={"/import/fruit"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Fruit"}</MenuItem>
            <MenuItem href={"/import/fruitingChamber"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Fruiting Chamber"}</MenuItem>
            <MenuItem href={"/import/jar"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Jar"}</MenuItem>
            <MenuItem href={"/import/lc"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Liquid Culture"}</MenuItem>
            <MenuItem href={"/import/lcSyringe"} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Liquid Culture Syringe"}</MenuItem>
            <MenuItem href={"/import/mss"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Multi-Spore Syringe"}</MenuItem>
            <MenuItem href={"/import/plate"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Plate"}</MenuItem>
            <MenuItem href={"/import/plugs"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Plugs"}</MenuItem>
            <MenuItem href={"/import/slant"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Slant"}</MenuItem>
            <MenuItem href={"/import/sporePrint"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Spore Print"}</MenuItem>
            <MenuItem href={"/import/stasisTube"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Stasis Tube"}</MenuItem>
        </Menu>
    </div>
}

export function TopBarListMenu() {
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl)
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorEl(event.currentTarget)
    }
    const handleClose = () => {
        setAnchorEl(null)
    }
    return <div>
        <Button
            id={"topBarListButton"}
            sx={buttonProps}// TODO: COLORS
            aria-controls={open ? 'topBarListMenu' : undefined}
            aria-haspopup={true}
            aria-expanded={open ? 'true' : undefined}
            onClick={handleClick}>
            {"List"}
        </Button>
        <Menu id={"topBarListMenu"}
              anchorEl={anchorEl}
              open={open}
              onClose={handleClose}
              slotProps={{
                  list: {'aria-labelledby': 'topBarListButton'}
              }}>

            <MenuItem href={"/list/agarBatches"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Agar Batches"}</MenuItem>
            <MenuItem href={"/list/agarRecipes"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Agar Recipes"}</MenuItem>
            <MenuItem href={"/list/bags"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Bags"}</MenuItem>
            <MenuItem href={"/list/fruits"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Fruits"}</MenuItem>
            <MenuItem href={"/list/fruitingChambers"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"FruitingChambers"}</MenuItem>
            <MenuItem href={"/list/grainBatches"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Grain Batches"}</MenuItem>
            <MenuItem href={"/list/jars"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Jars"}</MenuItem>
            <MenuItem href={"/list/jarRecipes"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Jar Recipes"}</MenuItem>
            <MenuItem href={"/list/lcs"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Liquid Cultures"}</MenuItem>
            <MenuItem href={"/list/lcRecipes"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"Liquid Culture Recipes"}</MenuItem>
            <MenuItem href={"/list/lcSyringes"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"Liquid Culture Syringes"}</MenuItem>
            <MenuItem href={"/list/mss"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"MultiSpore Syringes"}</MenuItem>
            <MenuItem href={"/list/pcRuns"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"PcRuns"}</MenuItem>
            <MenuItem href={"/list/plates"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Plates"}</MenuItem>
            <MenuItem href={"/list/plugs"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Plugs"}</MenuItem>
            <MenuItem href={"/list/projects"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Projects"}</MenuItem>
            <MenuItem href={"/list/sales"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Sales"}</MenuItem>
            <MenuItem href={"/list/slants"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Slants"}</MenuItem>
            <MenuItem href={"/list/species"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Species"}</MenuItem>
            <MenuItem href={"/list/sporePrints"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Spore Prints"}</MenuItem>
            <MenuItem href={"/list/sporeSwabs"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Spore Swabs"}</MenuItem>
            <MenuItem href={"/list/stasisTubes"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Stasis Tubes"}</MenuItem>
            <MenuItem href={"/list/subspecies"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Subspecies"}</MenuItem>
            <MenuItem href={"/list/substrateBatches"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"Substrate Batches"}</MenuItem>
            <MenuItem href={"/list/substrateRecipes"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"Substrate Recipes"}</MenuItem>
            <MenuItem href={"/list/transfers"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Transfers"}</MenuItem>
            <MenuItem href={"/list/users"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Users"}</MenuItem>
            <MenuItem href={"/list/waterJars"} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Water Jars"}</MenuItem>
        </Menu>
    </div>
}