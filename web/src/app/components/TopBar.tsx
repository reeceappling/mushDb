"use client"

import ReaderWriterSelector, {
    ReadTagFunc,
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import * as React from "react";
import {JSX, useState} from "react";
import Button from "@mui/material/Button"
import Menu from "@mui/material/Menu"
import MenuItem from "@mui/material/MenuItem"
import TextBox from "@/app/components/formSubcomponents/textbox";
import {getPathFor, webUrl} from "@/app/components/common";


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
    const menuItem = (entryType:string, txt:string):JSX.Element => {
        return <MenuItem href={"/new/"+entryType} onClick={handleClose} component={"a"} sx={sublistItemProps}>{txt}</MenuItem>
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
            {menuItem("agarRecipe", "Agar Recipe")}
            {menuItem("jarRecipe", "Jar Recipe")}
            {menuItem("lcRecipe", "LC Recipe")}
            {menuItem("pcRun", "PC Run")}{/* TODO: PC RUN??? */}
            {menuItem("project", "Project")}{/* TODO: maybe just create this in each form? */}
            {menuItem("species", "Species")}
            {menuItem("subspecies", "Subspecies")}
            {menuItem("substrateRecipe", "Substrate Recipe")}
            {menuItem("waterJar", "Water Jar")}{/* TODO: ALSO ALLOW IT TO BE DONE FROM THE PC RUN PAGE*/}
        </Menu>
    </div>
}

export default function TopBar() {
    // TODO: RECENTS FOR ALL ENTRIES?????
    const {dispatch} = useRfidReaderContext()
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
        <TopBarViewMenu/> {/* TODO: ENSURE LINKS ARE CORRECT!!!*/}
        <TopBarImportMenu/>
        <TopBarCreateMenu/>
        <div id={"rfidTopArea"}>
            <LastReadTag/>
            <ReadTagButton/>
            <ReaderWriterSelector onSelect={onReaderSelect}/>
        </div>
    </div>
}

// function CopyLatestReadTagButton() {
//     const {state, dispatch} = useRfidReaderContext()
//     if (state.lastReadTag == undefined) {
//         return null
//     }
//     const onClick = () => {
//         if (state.lastReadTag != undefined) {
//             navigator.clipboard.writeText(state.lastReadTag).catch((err) => {
//                 let toWrite = "failed to copy tag value to clipboard: " + err
//                 console.error(toWrite)
//                 dispatch({
//                     type: ActionTypes.SET_ERROR,
//                     payload: toWrite,
//                 })
//             })
//         }
//     }
//     return <button className={"basicButtonSmall"} onClick={onClick}>{"Copy last read tag value"}</button>
// }

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
    return <button className={"basicButtonSmall"} onClick={onButtonClick}>{"Use latest read tag"}</button>
}


function ReadTagButton({onResult}: { onResult?: (id: string) => void }) {
    const {state, dispatch} = useRfidReaderContext()
    const onClick = () => {
        if (state.selected != undefined) {
            //const a = ReadRfidTag(state.selected) // TODO: REENABLE IF/WHEN WE CAN!
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
        getPathFor(id).then((path) => {
            location.assign(webUrl("/view/" + path))
        }).catch((err) => {
            // TODO: handle the error!
            console.log("failed to get path for id: "+JSON.stringify(err))
        })
    }
    const {state} = useRfidReaderContext()
    const redirectForId = (redirectToId:string)=>{
        setId(redirectToId)
        handleViewById()
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
            <MenuItem onClick={(e) => {
                e.preventDefault(); // TODO: ensure ok
                e.stopPropagation();
            }}>
                <div>
                    <div>{"Main Collection Item By ID"}</div>
                    <TextBox readonly={false} label={"ID"} value={id} fieldName={"viewByIdInput"}
                             updateTextHandler={setId}/>
                    <ReadTagButton onResult={redirectForId}/>
                    {state.lastReadTag && <UseLatestReadTagButton onClick={(v) => {
                        v && setId(v)
                    }}/>}
                    {id !== "" && <button className={"greenButton buttonSmall"} onClick={e=>{
                        e.stopPropagation();
                        handleViewById()
                    }}> {"go to this id"}</button>}
                </div>
            </MenuItem> {/* TODO: ENSURE EACH LINK WORKS*/}
            <MenuItem href={"/testpage"} onClick={handleClose}
                      component={"a"} sx={sublistItemProps}>{"TEST PAGE"}</MenuItem>{/*TODO: DELETE ME*/}
            <MenuItem href={"/view/agarBatch/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Agar Batch"}</MenuItem>
            <MenuItem href={"/view/agarRecipe/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Agar Recipe"}</MenuItem>
            <MenuItem href={"/view/jarRecipe/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Jar Recipe"}</MenuItem>
            <MenuItem href={"/view/lcRecipe/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Liquid Culture Recipe"}</MenuItem>
            <MenuItem href={"/view/pcRun/"+id} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"PC Run"}</MenuItem>
            <MenuItem href={"/view/project/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Project"}</MenuItem>
            <MenuItem href={"/view/sale/"+id} onClick={handleClose} component={"a"} sx={sublistItemProps}>{"Sale"}</MenuItem>
            <MenuItem href={"/view/species/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Species"}</MenuItem>
            <MenuItem href={"/view/subspecies/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Subspecies"}</MenuItem>
            <MenuItem href={"/view/substrateBatch/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Substrate Batch"}</MenuItem>
            <MenuItem href={"/view/substrateRecipe/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"Substrate Recipe"}</MenuItem>
            <MenuItem href={"/view/transfer/"+id} onClick={handleClose}
                       component={"a"} sx={sublistItemProps}>{"Transfer"}</MenuItem>
            <MenuItem href={"/view/user/"+id} onClick={handleClose} component={"a"} sx={sublistItemProps}>{/* TODO: BY NAME? urlencode???*/"User"}</MenuItem>
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
    const menuItem = (path: string, txt: string):JSX.Element => {
        return <MenuItem onClick={handleClose} component={"a"} sx={sublistItemProps} href={path} >{txt}</MenuItem>
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
            {menuItem("/import/bag", "Bag")}
            {menuItem("/import/fruit", "Fruit")}
            {menuItem("/import/fruitingChamber", "Fruiting Chamber")}
            {menuItem("/import/jar", "Jar")}
            {menuItem("/import/lc", "Liquid Culture")}
            {menuItem("/import/lcSyringe", "Liquid Culture Syringe")}
            {menuItem("/import/mss", "Multi-Spore Syringe")}
            {menuItem("/import/plate", "Plate")}
            {menuItem("/import/plugs", "Plugs")}
            {menuItem("/import/slant", "Slant")}
            {menuItem("/import/sporePrint", "Spore Print")}
            {menuItem("/import/stasisTube", "Stasis Tube")}
            {/* TODO: ?menuItem("/import/waterJar", "Water Jar")*/}
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
    const menuItem = (entryType: string, txt: string):JSX.Element => {
        return <MenuItem href={"/list/"+entryType} onClick={handleClose} component={"a"} sx={sublistItemProps}>{txt}</MenuItem>
    }
    return <div>
        <Button
            id={"topBarListButton"}
            sx={buttonProps}
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
            {menuItem("agarBatches", "Agar Batches")}
            {menuItem("agarRecipes", "Agar Recipes")}
            {menuItem("bags", "Bags")}
            {menuItem("fruits", "Fruits")}
            {menuItem("fruitingChambers", "Fruiting Chambers")}
            {menuItem("grainBatches", "Grain Batches")}
            {menuItem("jars", "Jars")}
            {menuItem("jarRecipes", "Jar Recipes")}
            {menuItem("lcs", "Liquid Cultures")}
            {menuItem("lcRecipes", "Liquid Culture Recipes")}
            {menuItem("lcSyringes", "Liquid Culture Syringes")}
            {menuItem("mss", "MultiSpore Syringes")}
            {menuItem("pcRuns", "PcRuns")}
            {menuItem("plates", "Plates")}
            {menuItem("plugs", "Plugs")}
            {menuItem("projects", "Projects")}
            {menuItem("sales", "Sales")}
            {menuItem("slants", "Slants")}
            {menuItem("species", "Species")}
            {menuItem("sporePrints", "Spore Prints")}
            {menuItem("sporeSwabs", "Spore Swabs")}
            {menuItem("stasisTubes", "Stasis Tubes")}
            {menuItem("subspecies", "Subspecies")}
            {menuItem("substrateBatches", "Substrate Batches")}
            {menuItem("substrateRecipes", "Substrate Recipes")}
            {menuItem("transfers", "Transfers")}
            {menuItem("users", "Users")}
            {menuItem("waterJars", "Water Jars")}
        </Menu>
    </div>
}